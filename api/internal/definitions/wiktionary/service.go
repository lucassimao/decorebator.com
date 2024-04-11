package wiktionary

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"decorebator.com/internal/common"
	"decorebator.com/internal/definitions"
	"decorebator.com/internal/definitions/openai"
)

// Type alias for convenience
type Definition = definitions.Definition

type WiktionaryExample struct {
	Text string `json:"text"`
}

type WiktionarySense struct {
	Glosses  []string            `json:"glosses"`
	Examples []WiktionaryExample `json:"examples"`
}
type WiktionaryData struct {
	PartOfSpeech string            `json:"pos"`
	Language     string            `json:"lang"`
	Word         string            `json:"word"`
	Senses       []WiktionarySense `json:"senses"`
}

func GetDefinition(token string) ([]Definition, error) {
	db, err := common.GetDBConnection()
	if err != nil {
		return nil, fmt.Errorf("could not open db connection: %w", err)
	}

	rows, err := db.Query(context.Background(), "SELECT id,word,data FROM wiktionary WHERE word = $1", token)
	if err != nil {
		return nil, fmt.Errorf("error reading from wiktionary: %w", err)
	}

	var definitions []Definition
	for rows.Next() {
		var id int
		var token string
		var data []byte
		var jsonData WiktionaryData

		err = rows.Scan(&id, &token, &data)
		if err != nil {
			return nil, fmt.Errorf("error scanning db result row: %w", err)
		}

		if err := json.Unmarshal(data, &jsonData); err != nil {
			return nil, fmt.Errorf("error unmarshal wikionary data: %w", err)
		}

		for _, sense := range jsonData.Senses {
			var glosses = sense.Glosses
			var examples []string

			for _, example := range sense.Examples {
				examples = append(examples, example.Text)
			}

			// making sure that each definition has at least 5 examples
			if len(examples) < 5 {
				additionalExamples, err := openai.GetExamples(token, jsonData.PartOfSpeech, 5-len(examples), glosses[0])

				if err != nil {
					log.Fatalf("Error getting examples from OpenAI: %v", err)
				} else {
					examples = append(examples, additionalExamples...)
				}
			}

			fmt.Printf("Examples: %v\n", examples)

			var definition Definition
			definition.Token = token
			definition.Meaning = glosses[0]
			definition.Examples = examples
			definition.PartOfSpeech = jsonData.PartOfSpeech
			definition.Language = jsonData.Language
			definition.Source = "wiktionary"

			definitions = append(definitions, definition)
		}
	}

	return definitions, nil
}
