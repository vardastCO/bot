package main

import (
	"log"
    "os/exec"
    "sync"
)

func main() {
	// Define a list of websites to scrape concurrently.
	websites := []string{
		"https://example.com",
		"https://example2.com",
		// Add more websites to scrape.
	}

	// Define the number of worker Goroutines.
	numWorkers := 100

	// Create a wait group to wait for all Goroutines to finish.
	var wg sync.WaitGroup

	// Create a channel to distribute the websites among worker Goroutines.
	websiteChannel := make(chan string, len(websites))

	// Fill the channel with websites.
	for _, website := range websites {
		websiteChannel <- website
	}
	close(websiteChannel) // Close the channel to signal that all websites are enqueued.

	// Create worker Goroutines.
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for website := range websiteChannel {
				// Run the scrape.js file using Node.js.
				cmd := exec.Command("node", "scrape.js", website)

				// Set the working directory to the location of your scrape.js file.
				cmd.Dir = "././node-app"

				// Capture the output (stdout and stderr) of the Node.js process.
				output, err := cmd.CombinedOutput()
				if err != nil {
					log.Printf("Error executing scrape.js for %s: %v\n%s", website, err, string(output))
				} else {
					log.Printf("Scrape.js executed successfully for %s:\n%s", website, string(output))
				}
			}
		}()
	}

	// Wait for all worker Goroutines to finish.
	wg.Wait()
}
