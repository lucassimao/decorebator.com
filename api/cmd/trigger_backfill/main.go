package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"decorebator.com/internal/service"
)

func main() {
	batchSize := 5 // Default batch size
	
	if len(os.Args) > 1 {
		if size, err := strconv.Atoi(os.Args[1]); err == nil && size > 0 && size <= 20 {
			batchSize = size
		}
	}

	fmt.Printf("Triggering backfill inflections job with batch size: %d\n", batchSize)

	result, err := service.TriggerBackfillInflectionsWorker(batchSize)
	if err != nil {
		log.Fatal("Failed to trigger backfill job:", err)
	}

	fmt.Printf("✓ Backfill job queued successfully! Job ID: %d\n", result.Job.ID)
	fmt.Println("\nThe job will:")
	fmt.Println("1. Process verb definitions in batches")
	fmt.Println("2. Generate inflections using ChatGPT")
	fmt.Println("3. Update existing definitions with inflection data")
	fmt.Println("4. Automatically queue the next batch if more work remains")
	fmt.Println("\nMonitor the worker logs to see progress.")
}