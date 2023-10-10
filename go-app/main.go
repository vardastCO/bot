package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Define a struct to represent the scraped data.
type ScrapeResult struct {
	URL   string `json:"url"`
	Price string `json:"price"`
}

var (
	dbPool    *pgxpool.Pool
	csvMutex  sync.Mutex
	csvFile   *os.File
	csvWriter *bufio.Writer
	wg        sync.WaitGroup // Define a WaitGroup for goroutine synchronization
)

func main() {
	// Initialize the database connection pool.
	dbConnStr := "postgresql://db:root@postgres:5432/db"

	poolConfig, err := pgxpool.ParseConfig(dbConnStr)
	if err != nil {
		log.Fatalf("Failed to parse DB connection string: %v", err)
	}
	dbPool, err = pgxpool.ConnectConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer dbPool.Close()

	// Create the CSV file for writing.
	createCSVFile()
	defer closeCSVFile()

	// Start one worker with the initial URL.
	initialURL := "https://hypersaz.com" // Replace with your initial URL
	startWorker(initialURL)

	// Create additional worker Goroutines.
	numWorkers := 2
	for i := 1; i < numWorkers; i++ {
		wg.Add(1) // Increment the WaitGroup for each worker
		go func(workerID int) {
			defer wg.Done()                    // Decrement the WaitGroup when the worker finishes
			workerURL := getURLToScrape()      // Implement a function to get the next URL to scrape
			startWorker(workerURL)
		}(i)
	}

	// Wait for all worker Goroutines to finish.
	wg.Wait()
}

func createCSVFile() {
	// Open the CSV file for writing.
	var err error
	csvFile, err = os.Create("output.csv")
	if err != nil {
		log.Printf("Error creating CSV file: %v", err)
		return
	}

	csvWriter = bufio.NewWriter(csvFile)

	// Write the CSV header.
	csvWriter.WriteString("URL,Price\n")
	csvWriter.Flush()
}

func closeCSVFile() {
	if csvFile != nil {
		csvFile.Close()
	}
}

func startWorker(url string) {
	// Create a worker Goroutine.
	go func() {
		defer wg.Done() // Decrement the WaitGroup when the worker finishes

		// Capture the JSON output from scrape.js for this worker.
		cmd := exec.Command("node", "scrape.js", url)
		cmd.Env = append(os.Environ(), "WORKER_ID=1") // You can set a unique worker ID here

		// Capture stdout from the Node.js process.
		output, err := cmd.StdoutPipe()
		if err != nil {
			log.Printf("Error getting stdout pipe for worker: %v", err)
			return
		}

		// Start the Node.js process.
		if err := cmd.Start(); err != nil {
			log.Printf("Error executing scrape.js for worker: %v", err)
			return
		}

		// Create a scanner to read the JSON data from stdout.
		scanner := bufio.NewScanner(output)

		for scanner.Scan() {
			jsonStr := scanner.Text()

			// Parse JSON data into a struct.
			var result ScrapeResult
			if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
				log.Printf("Error decoding JSON data from scrape.js for worker: %v", err)
				continue
			}

			// Handle the scraped data as needed (e.g., store it in a database).
			handleScrapedData(result)
		}

		if err := scanner.Err(); err != nil {
			log.Printf("Error reading JSON data from scrape.js for worker: %v", err)
		}

		// Wait for the Node.js process to finish.
		if err := cmd.Wait(); err != nil {
			log.Printf("Error waiting for scrape.js for worker: %v", err)
		}
	}()
}

func handleScrapedData(data ScrapeResult) {
	// Handle the scraped data here, e.g., store it in a database.

	// Write the data to the CSV file.
	csvMutex.Lock()
	defer csvMutex.Unlock()

	if csvWriter != nil {
		// Format the data as a CSV row.
		csvRow := fmt.Sprintf("%s,%s\n", data.URL, data.Price)

		// Write the CSV row to the file.
		csvWriter.WriteString(csvRow)
		csvWriter.Flush()
	}

	// After processing the data, move the URL to the visited table in the database.
	markURLAsVisited(data.URL)
}

func markURLAsVisited(url string) {
	_, err := dbPool.Exec(context.Background(), "INSERT INTO visited (url) VALUES ($1)", url)
	if err != nil {
		log.Printf("Error marking URL as visited: %v", err)
	}
}

func getURLToScrape() string {
	// Implement a function to get the next URL to scrape from the unvisited table.
	var url string
	err := dbPool.QueryRow(context.Background(), "SELECT url FROM unvisited LIMIT 1").Scan(&url)
	if err != nil {
		if err == pgx.ErrNoRows {
			// No unvisited URLs left to scrape.
			return ""
		}
		log.Printf("Error getting URL to scrape: %v", err)
		return ""
	}

	// Delete the URL from the unvisited table.
	_, err = dbPool.Exec(context.Background(), "DELETE FROM unvisited WHERE url = $1", url)
	if err != nil {
		log.Printf("Error deleting URL from unvisited table: %v", err)
	}

	return url
}
