const puppeteer = require('puppeteer-core');
const puppeteerExtra = require('puppeteer-extra');
const ProxyPlugin = require('puppeteer-extra-plugin-proxy');

puppeteerExtra.use(ProxyPlugin());
const url = process.argv[2]; // Get the URL from the command-line arguments.

if (!url) {
    console.error('Usage: node scrape.js <URL>');
    process.exit(1);
}

(async () => {
  try {
    const proxyServer = 'ss://YWVzLTI1Ni1nY206d0dVaGt6WGpjRA==@38.54.13.15:31214#main'; // Replace with your Shadowsocks proxy URL

    const browser = await puppeteer.launch({
      headless: false,
      executablePath: 'C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe',
    });

    const page = await browser.newPage();

    // Configure the Shadowsocks proxy using the plugin.
    await page.authenticate({
      username: proxyServer,
    });

    // Log a message to indicate that the proxy is being used.
    // console.log('Using Proxy:', proxyServer);
    //  await page.goto('https://whatismyipaddress.com/'); // Replace with the URL of the website you want to scrape.

    await page.goto(url); // Replace with the URL of the website you want to scrape.
    // Use XPath to select the desired element.
    const xpathExpression = '/html/body/div[1]/div[1]/div[2]/div[4]/div[2]/div[2]/div[2]/div[2]/div[2]/div[2]/ul/li[2]/p';
    const [element] = await page.$x(xpathExpression);

    if (element) {
      // Extract the text content of the selected element.
      const elementText = await page.evaluate(el => el.textContent, element);
      console.log('Selected element text:', elementText);
    } else {
      console.log('Element not found');
    }

    // Close the browser when done.
    await browser.close();
  } catch (error) {
    console.error('An error occurred:', error);
  }
})();

