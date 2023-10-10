const puppeteer = require('puppeteer');
const { Client } = require('pg');
require('dotenv').config();

const url = process.argv[2]; // Get the URL from the command-line arguments.

if (!url) {
    console.error('Usage: node scrape.js <URL>');
    process.exit(1);
}

const pool = new Client({
  user: 'db', // PostgreSQL username
  host: 'postgres', // Use the service name 'postgres' as the host
  database: 'db', // PostgreSQL database name
  password: 'root', // PostgreSQL password
  port: 5432, // PostgreSQL port
});

const initialPage = 'https://www.hypersaz.com/';
const startUrlPattern = 'https://www.hypersaz.com/product.php?';

async function createBrowser() {
    try {
        const browser = await puppeteer.launch({
            headless: true,
        });
        return browser;
    } catch (error) {
        console.error('Error creating the browser:', error);
        throw error;
    }
}

const processedHrefs = new Set();
const unprocessedHrefs = new Set();

async function processPage(pageUrl) {
  console.log('Processing Page:', pageUrl);
  try {
      await pool.connect();
      const visitedResult = await pool.query('SELECT id FROM visited WHERE url = $1', [pageUrl]);

      if (visitedResult.rows.length === 0) {
          await pool.query('INSERT INTO visited(url) VALUES($1)', [pageUrl]);

          const page = await browser.newPage();
          await page.goto(pageUrl, { timeout: 300000 });
      
          const priceElement = await page.$x(
              '/html/body/section[2]/div/div/div[3]/div/ul/li[2]/p/span'
          );
          if (priceElement.length > 0) {
              const priceText = await page.evaluate(
                  (el) => el.textContent,
                  priceElement[0]
              );
              if (priceText.trim() !== '') {
            
                await pool.query('INSERT INTO scraped_data(url, price) VALUES($1, $2)', [pageUrl, priceText.trim()]);
                console.log(`Saved: URL: ${pageUrl}, Price: ${priceText.trim()}`);
            }
          }

          const hrefs = await page.evaluate(() => {
              const links = Array.from(document.querySelectorAll('a'));
              return links.map((link) => link.getAttribute('href'));
          });

          for (const href of hrefs) {
              if (href && !processedHrefs.has(href)) {
                  if (!href.startsWith('https://')) {
                      var outputUrl = initialPage + href;
                  } else {
                      var outputUrl = href;
                  }
                  if (outputUrl.startsWith(startUrlPattern)) {
                      // Insert the URL into the 'unvisited' table
                      await pool.query('INSERT INTO unvisited(url) VALUES($1)', [outputUrl]);
                      unprocessedHrefs.add(outputUrl);
                  }
              }
          }

          await page.close();
      }
  } catch (error) {
      console.error('An error occurred while navigating to the page:', error);
      // Handle the error as needed
  } finally {
      await pool.end(); // Disconnect from the PostgreSQL database
  }
}


async function main() {
    const browser = await createBrowser();

    try {
        unprocessedHrefs.add(initialPage);

        while (unprocessedHrefs.size > 0) {
            const currentHref = Array.from(unprocessedHrefs)[0];
            unprocessedHrefs.delete(currentHref);
            processedHrefs.add(currentHref);

            await processPage(currentHref);
        }
    } catch (error) {
        console.error('An error occurred:', error);
    } finally {
        if (browser) {
            await browser.close();
        }
        console.log('Finished scraping.');
    }
}

createBrowser()
    .then(() => {
        main();
    })
    .catch(() => {
        console.log('Failed to create a browser.');
    });
