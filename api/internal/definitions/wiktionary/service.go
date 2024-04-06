package wiktionary

import (
	"context"
	"encoding/json"
	"fmt"

	"decorebator.com/internal/common"
	"decorebator.com/internal/definitions"
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
		return nil, err
	}

	rows, err := db.Query(context.Background(), "SELECT id,word,data FROM wiktionary WHERE word = $1", token)
	if err != nil {
		return nil, err
	}

	var definitions []Definition
	for rows.Next() {
		var id int
		var token string
		var data []byte
		var jsonData WiktionaryData

		err = rows.Scan(&id, &token, &data)
		if err != nil {
			return nil, err
		}

		// Unmarshal the jsonData into the map
		if err := json.Unmarshal(data, &jsonData); err != nil {
			return nil, err
		}

		for _, sense := range jsonData.Senses {
			var glosses = sense.Glosses
			fmt.Printf("Glosses: %v\n", glosses)

			var examples []string
			for _, example := range sense.Examples {
				examples = append(examples, example.Text)
			}

			fmt.Printf("Examples: %v\n", examples)
			// use openapi to fetch examples if needed taking into account the part of speech and word

			var definition Definition
			definition.Meaning = glosses[0]
			definition.Examples = examples
			definition.PartOfSpeech = jsonData.PartOfSpeech
			definition.Language = jsonData.Language

			definitions = append(definitions, definition)
		}
	}

	return definitions, nil
}
