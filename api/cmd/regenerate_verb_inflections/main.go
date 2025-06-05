package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"decorebator.com/internal/common"
	"decorebator.com/internal/service"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/regenerate_verb_inflections/main.go <batch_size>")
		fmt.Println("Example: go run cmd/regenerate_verb_inflections/main.go 10")
		os.Exit(1)
	}

	batchSize, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal("Invalid batch size:", err)
	}

	if batchSize <= 0 || batchSize > 100 {
		log.Fatal("Batch size must be between 1 and 100")
	}

	fmt.Printf("Starting regeneration of verb inflections in batches of %d...\n", batchSize)

	db, err := common.GetDBConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Find words with verb definitions that have no inflections
	query := `
		SELECT DISTINCT w.id, w.name
		FROM words w
		JOIN word_definitions wd ON wd.word_id = w.id  
		JOIN definitions d ON d.id = wd.definition_id
		WHERE d.part_of_speech IN ('verb', 'phrasal verb')
		AND (d.inflections IS NULL OR jsonb_array_length(d.inflections) = 0)
		ORDER BY w.id
		LIMIT $1
	`

	rows, err := db.Query(context.Background(), query, batchSize)
	if err != nil {
		log.Fatal("Query failed:", err)
	}
	defer rows.Close()

	var wordsToProcess []struct {
		ID   int64
		Name string
	}

	for rows.Next() {
		var id int64
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		wordsToProcess = append(wordsToProcess, struct {
			ID   int64
			Name string
		}{id, name})
	}

	if len(wordsToProcess) == 0 {
		fmt.Println("No words found that need inflection regeneration.")
		return
	}

	fmt.Printf("Found %d words that need inflection regeneration:\n", len(wordsToProcess))
	for _, word := range wordsToProcess {
		fmt.Printf("- ID %d: %s\n", word.ID, word.Name)
	}

	fmt.Print("\nProceed with regeneration? (y/n): ")
	var response string
	fmt.Scanln(&response)
	
	if response != "y" && response != "Y" {
		fmt.Println("Aborted.")
		return
	}

	// Process each word
	for i, word := range wordsToProcess {
		fmt.Printf("[%d/%d] Processing word: %s (ID: %d)\n", i+1, len(wordsToProcess), word.Name, word.ID)

		// Delete existing definitions for this word
		tx, err := db.Begin(context.Background())
		if err != nil {
			log.Printf("Failed to start transaction for word %s: %v", word.Name, err)
			continue
		}

		// Delete existing verb/phrasal verb definitions that don't have inflections
		_, err = tx.Exec(context.Background(), `
			DELETE FROM definitions 
			WHERE id IN (
				SELECT d.id 
				FROM definitions d
				JOIN word_definitions wd ON wd.definition_id = d.id
				WHERE wd.word_id = $1 
				AND d.part_of_speech IN ('verb', 'phrasal verb')
				AND (d.inflections IS NULL OR jsonb_array_length(d.inflections) = 0)
			)
		`, word.ID)

		if err != nil {
			log.Printf("Failed to delete old definitions for word %s: %v", word.Name, err)
			tx.Rollback(context.Background())
			continue
		}

		// Trigger the definition fetcher worker for this word
		_, err = service.TriggerFetchDefinitionWorker(word.ID, nil, &tx)
		if err != nil {
			log.Printf("Failed to trigger definition fetch for word %s: %v", word.Name, err)
			tx.Rollback(context.Background())
			continue
		}

		err = tx.Commit(context.Background())
		if err != nil {
			log.Printf("Failed to commit transaction for word %s: %v", word.Name, err)
			continue
		}

		fmt.Printf("✓ Successfully queued regeneration for: %s\n", word.Name)
	}

	fmt.Printf("\nCompleted! Processed %d words.\n", len(wordsToProcess))
	fmt.Println("Note: The definition fetcher workers will run asynchronously to generate new definitions with inflections.")
	fmt.Println("You can run this script again to process more words in batches.")
}