package wiktionary

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"decorebator.com/internal/common"
	"decorebator.com/internal/definitions"
	"decorebator.com/internal/definitions/openai"
	"github.com/jackc/pgx/v5"
)

// Type alias for convenience
type Definition = definitions.Definition

type WiktionaryExample struct {
	Text string `json:"text"`
}

type FormOf struct {
	Word string `json:"word"`
}

type WiktionarySense struct {
	Glosses  []string            `json:"glosses"`
	Examples []WiktionaryExample `json:"examples"`
	FormOf   []FormOf            `json:"form_of"`
}
type WiktionaryData struct {
	PartOfSpeech string            `json:"pos"`
	Language     string            `json:"lang"`
	Word         string            `json:"word"`
	Senses       []WiktionarySense `json:"senses"`
}

func GetDefinition(token string) ([]Definition, error) {
	log.Printf("searching %s definition in wiktionary\n", token)

	db, err := common.GetDBConnection()
	if err != nil {
		return nil, fmt.Errorf("could not open db connection: %w", err)
	}

	rows, err := db.Query(context.Background(), "SELECT id,word,data FROM wiktionary WHERE word = $1", token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			fmt.Printf("no definition for %s in wiktionary\n", token)
			return nil, nil
		}
		return nil, fmt.Errorf("error reading from wiktionary: %w", err)
	}
	defer rows.Close()

	var definitions []Definition
	for rows.Next() {
		var id int
		var token string
		var data []byte
		var jsonData WiktionaryData

		err = rows.Scan(&id, &token, &data)

		if err != nil {
			return nil, fmt.Errorf("error scanning db result for word %s: %w", token, err)
		}

		if err := json.Unmarshal(data, &jsonData); err != nil {
			return nil, fmt.Errorf("error unmarshaling wikionary for word %s data: %w", token, err)
		}

		for _, sense := range jsonData.Senses {
			var glosses = sense.Glosses
			var examples []string

			// if the token is any form of a verb, then get the definition of the verb instead
			if len(sense.FormOf) > 0 && jsonData.PartOfSpeech == "verb" {
				verbDefinitions, err := GetDefinition(sense.FormOf[0].Word)
				if err != nil {
					return nil, fmt.Errorf("error searching %s definition: %w", sense.FormOf[0].Word, err)
				}
				definitions = append(definitions, verbDefinitions...)
				continue
			}

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

	log.Printf("%d definitions for %s found in wiktionary\n", len(definitions), token)
	return definitions, nil
}
