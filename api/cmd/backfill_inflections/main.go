package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"decorebator.com/internal/common"
	"decorebator.com/internal/openai"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/backfill_inflections/main.go <batch_size>")
		fmt.Println("Example: go run cmd/backfill_inflections/main.go 5")
		fmt.Println("Note: Keep batch size small to respect OpenAI API rate limits")
		os.Exit(1)
	}

	batchSize, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal("Invalid batch size:", err)
	}

	if batchSize <= 0 || batchSize > 20 {
		log.Fatal("Batch size must be between 1 and 20 (to respect API limits)")
	}

	fmt.Printf("Starting backfill of missing inflections in batches of %d...\n", batchSize)

	db, err := common.GetDBConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Find verb definitions that don't have inflections
	query := `
		SELECT id, token, part_of_speech
		FROM definitions
		WHERE part_of_speech IN ('verb', 'phrasal verb')
		AND (inflections IS NULL OR jsonb_array_length(inflections) = 0)
		ORDER BY id
		LIMIT $1
	`

	rows, err := db.Query(context.Background(), query, batchSize)
	if err != nil {
		log.Fatal("Query failed:", err)
	}
	defer rows.Close()

	var definitionsToProcess []struct {
		ID           int64
		Token        string
		PartOfSpeech string
	}

	for rows.Next() {
		var id int64
		var token, pos string
		err = rows.Scan(&id, &token, &pos)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		definitionsToProcess = append(definitionsToProcess, struct {
			ID           int64
			Token        string
			PartOfSpeech string
		}{id, token, pos})
	}

	if len(definitionsToProcess) == 0 {
		fmt.Println("No verb definitions found that need inflection backfill.")
		return
	}

	fmt.Printf("Found %d verb definitions that need inflection backfill:\n", len(definitionsToProcess))
	for _, def := range definitionsToProcess {
		fmt.Printf("- ID %d: %s (%s)\n", def.ID, def.Token, def.PartOfSpeech)
	}

	fmt.Print("\nProceed with backfill? (y/n): ")
	var response string
	fmt.Scanln(&response)
	
	if response != "y" && response != "Y" {
		fmt.Println("Aborted.")
		return
	}

	// Group definitions by token to avoid duplicate API calls
	tokenMap := make(map[string][]int64)
	for _, def := range definitionsToProcess {
		tokenMap[def.Token] = append(tokenMap[def.Token], def.ID)
	}

	successCount := 0
	errorCount := 0

	// Process each unique token
	for token, definitionIds := range tokenMap {
		fmt.Printf("Processing token: %s (affects %d definitions)\n", token, len(definitionIds))

		// Get new definition with inflections from OpenAI
		definitionData, err := openai.GetDefinition(token)
		if err != nil {
			log.Printf("Failed to get definition for '%s': %v", token, err)
			errorCount++
			continue
		}

		// Find verb definitions with inflections in the response
		var verbDefinitionsWithInflections []struct {
			PartOfSpeech string
			Inflections  []struct {
				Inflection string   `json:"inflection"`
				Tense      string   `json:"tense"`
				Examples   []string `json:"examples"`
			}
		}

		for _, def := range definitionData.Definitions {
			if (def.PartOfSpeech == "verb" || def.PartOfSpeech == "phrasal verb") && len(def.Inflections) > 0 {
				verbDefinitionsWithInflections = append(verbDefinitionsWithInflections, struct {
					PartOfSpeech string
					Inflections  []struct {
						Inflection string   `json:"inflection"`
						Tense      string   `json:"tense"`
						Examples   []string `json:"examples"`
					}
				}{
					PartOfSpeech: def.PartOfSpeech,
					Inflections:  def.Inflections,
				})
			}
		}

		if len(verbDefinitionsWithInflections) == 0 {
			log.Printf("No verb definitions with inflections found for '%s'", token)
			errorCount++
			continue
		}

		// Update existing definitions with inflections
		for _, defId := range definitionIds {
			// Use the first verb definition's inflections (they should be the same for the same word)
			inflections := verbDefinitionsWithInflections[0].Inflections

			updateQuery := `
				UPDATE definitions 
				SET inflections = $1,
				    updated_at = NOW()
				WHERE id = $2
			`

			_, err = db.Exec(context.Background(), updateQuery, inflections, defId)
			if err != nil {
				log.Printf("Failed to update definition ID %d: %v", defId, err)
				errorCount++
			} else {
				fmt.Printf("✓ Updated definition ID %d with %d inflections\n", defId, len(inflections))
				successCount++
			}
		}
	}

	fmt.Printf("\nBackfill completed!\n")
	fmt.Printf("Successfully updated: %d definitions\n", successCount)
	fmt.Printf("Errors: %d\n", errorCount)
	fmt.Println("\nYou can run this script again to process more definitions in batches.")
}