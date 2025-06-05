package main

import (
	"context"
	"fmt"
	"log"

	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
)

func main() {
	
	db, err := common.GetDBConnection()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Query definitions with inflections but empty part_of_speech
	query := `
		SELECT id, token, part_of_speech, inflections
		FROM definitions 
		WHERE inflections IS NOT NULL 
		AND jsonb_array_length(inflections) > 0
		AND (part_of_speech IS NULL OR part_of_speech = '')
		LIMIT 10
	`

	rows, err := db.Query(context.Background(), query)
	if err != nil {
		log.Fatal("Query failed:", err)
	}
	defer rows.Close()

	fmt.Println("Definitions with inflections but missing part_of_speech:")
	fmt.Println("======================================================")

	count := 0
	for rows.Next() {
		var id int64
		var token string
		var partOfSpeech *string
		var inflections []model.Inflection

		err = rows.Scan(&id, &token, &partOfSpeech, &inflections)
		if err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		posValue := "NULL"
		if partOfSpeech != nil {
			posValue = *partOfSpeech
		}

		fmt.Printf("ID: %d, Token: %s, PartOfSpeech: %s\n", id, token, posValue)
		
		if len(inflections) > 0 {
			fmt.Printf("  Inflections: %d found\n", len(inflections))
			for i, inf := range inflections {
				fmt.Printf("    [%d] %s (%s) - %d examples\n", i, inf.Inflection, inf.Tense, len(inf.Examples))
			}
		}
		fmt.Println()
		count++
	}

	if count == 0 {
		fmt.Println("No definitions found with inflections but missing part_of_speech")
	} else {
		fmt.Printf("Found %d problematic definitions\n", count)
	}

	// Check specific definition ID 130
	fmt.Println("\nChecking definition ID 130:")
	fmt.Println("===========================")
	
	specificQuery := `
		SELECT id, token, part_of_speech, inflections, examples
		FROM definitions 
		WHERE id = 130
	`

	specificRows, err := db.Query(context.Background(), specificQuery)
	if err != nil {
		log.Printf("Specific query failed: %v", err)
	} else {
		defer specificRows.Close()
		
		if specificRows.Next() {
			var id int64
			var token string
			var partOfSpeech *string
			var inflections []model.Inflection
			var examples []string

			err = specificRows.Scan(&id, &token, &partOfSpeech, &inflections, &examples)
			if err != nil {
				log.Printf("Error scanning specific row: %v", err)
			} else {
				posValue := "NULL"
				if partOfSpeech != nil {
					posValue = *partOfSpeech
				}

				fmt.Printf("ID: %d\n", id)
				fmt.Printf("Token: %s\n", token)
				fmt.Printf("PartOfSpeech: %s\n", posValue)
				fmt.Printf("Examples count: %d\n", len(examples))
				fmt.Printf("Inflections count: %d\n", len(inflections))
				
				if len(inflections) > 0 {
					fmt.Println("Inflections details:")
					for i, inf := range inflections {
						fmt.Printf("  [%d] %s (%s) - %d examples\n", i, inf.Inflection, inf.Tense, len(inf.Examples))
					}
				}
			}
		} else {
			fmt.Println("Definition ID 130 not found")
		}
	}

	// Search for "rig" token specifically
	fmt.Println("\nSearching for 'rig' definitions:")
	fmt.Println("=================================")
	
	rigQuery := `
		SELECT id, token, part_of_speech, inflections, examples
		FROM definitions 
		WHERE token = 'rig'
	`

	rigRows, err := db.Query(context.Background(), rigQuery)
	if err != nil {
		log.Printf("Rig query failed: %v", err)
	} else {
		defer rigRows.Close()
		
		rigCount := 0
		for rigRows.Next() {
			var id int64
			var token string
			var partOfSpeech *string
			var inflections []model.Inflection
			var examples []string

			err = rigRows.Scan(&id, &token, &partOfSpeech, &inflections, &examples)
			if err != nil {
				log.Printf("Error scanning rig row: %v", err)
				continue
			}

			posValue := "NULL"
			if partOfSpeech != nil {
				posValue = *partOfSpeech
			}

			fmt.Printf("ID: %d, Token: %s, PartOfSpeech: %s\n", id, token, posValue)
			fmt.Printf("  Examples: %d, Inflections: %d\n", len(examples), len(inflections))
			
			if len(inflections) > 0 {
				fmt.Println("  Inflection details:")
				for i, inf := range inflections {
					fmt.Printf("    [%d] %s (%s) - %d examples\n", i, inf.Inflection, inf.Tense, len(inf.Examples))
				}
			}
			fmt.Println()
			rigCount++
		}
		
		if rigCount == 0 {
			fmt.Println("No definitions found for 'rig'")
		}
	}

	// Also check what part_of_speech values we have
	fmt.Println("\nExisting part_of_speech values in database:")
	fmt.Println("==========================================")
	
	posQuery := `
		SELECT part_of_speech, COUNT(*) 
		FROM definitions 
		WHERE part_of_speech IS NOT NULL AND part_of_speech != ''
		GROUP BY part_of_speech 
		ORDER BY COUNT(*) DESC
	`
	
	posRows, err := db.Query(context.Background(), posQuery)
	if err != nil {
		log.Fatal("POS query failed:", err)
	}
	defer posRows.Close()

	for posRows.Next() {
		var pos string
		var count int
		err = posRows.Scan(&pos, &count)
		if err != nil {
			log.Printf("Error scanning POS row: %v", err)
			continue
		}
		fmt.Printf("%s: %d\n", pos, count)
	}
}