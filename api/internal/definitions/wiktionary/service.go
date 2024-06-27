package wiktionary

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"decorebator.com/internal/common"
	"decorebator.com/internal/definitions"
	"decorebator.com/internal/definitions/openai"
	"github.com/jackc/pgx/v5"
	goredis "github.com/redis/go-redis/v9"
)

// Type alias for convenience
type Definition = definitions.Definition
type PhoneticNotation = definitions.PhoneticNotation
type Sound = definitions.Sound
type Accent = definitions.Accent

const Wiktionary = definitions.Wiktionary
const US = definitions.US
const CA = definitions.CANADA
const UK = definitions.UK

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

// represents an individual sound entry in the JSON
type WiktionarySound struct {
	IPA    string   `json:"ipa,omitempty"`
	Rhymes string   `json:"rhymes,omitempty"`
	Tags   []string `json:"tags,omitempty"`
	Text   string   `json:"text,omitempty"`
	Audio  string   `json:"audio,omitempty"`
	MP3URL string   `json:"mp3_url,omitempty"`
	OGGURL string   `json:"ogg_url,omitempty"`
}

type WiktionaryData struct {
	PartOfSpeech string            `json:"pos"`
	Language     string            `json:"lang"`
	Word         string            `json:"word"`
	Senses       []WiktionarySense `json:"senses"`
	Sounds       []WiktionarySound `json:"sounds"`
}

var redisAddr = os.Getenv("REDIS_ADDR")
var redis = goredis.NewClient(&goredis.Options{
	Addr:     redisAddr,
	Password: "", // no password set
	DB:       0,  // use default DB
})

func GetDefinition(token string) ([]Definition, error) {

	var wiktionaryIds []string
	logger := common.Logger.With("token", token, "func", "GetDefinition")

	// extract definition ids from redis
	result, err := redis.HGet(context.Background(), "wikitionary", token).Result()
	if err == nil {
		wiktionaryIds = strings.Split(result, ",")
	}

	db, err := common.GetDBConnection()
	if err != nil {
		return nil, fmt.Errorf("could not open db connection: %w", err)
	}

	var rows pgx.Rows

	if wiktionaryIds != nil {
		logger.Debug("indexes found in redis", "indexes", wiktionaryIds)
		rows, err = db.Query(context.Background(), "SELECT id,word,data FROM wiktionary WHERE id = ANY($1)", wiktionaryIds)
	} else {
		logger.Debug("searching definition in db")
		rows, err = db.Query(context.Background(), "SELECT id,word,data FROM wiktionary WHERE word = $1", token)
	}

	defer rows.Close()

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Debug("no definition found in db", "token", token)
			return nil, nil
		} else {
			logger.Error("error reading from db", "error", err)
		}
		return nil, fmt.Errorf("error reading from wiktionary")
	}

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

			// if the token is any form of a verb, then get the definition of the verb too
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

			// fallback to chatgpt if no enough examples
			if len(examples) < 5 {
				additionalExamples, err := openai.GetExamples(token, jsonData.PartOfSpeech, 5, glosses[0])

				if err != nil {
					logger.Error("Error getting examples from OpenAI", "error", err)
				} else {
					examples = append(examples, additionalExamples...)
				}
			}

			logger.Debug("Examples collected", "count", len(examples))

			var definition Definition
			definition.Token = token
			definition.Meaning = glosses[0]
			definition.Examples = examples
			definition.PartOfSpeech = jsonData.PartOfSpeech
			definition.Language = jsonData.Language
			definition.Source = Wiktionary
			definition.SourceId = strconv.Itoa(id)

			for _, v := range jsonData.Sounds {
				logger.Debug("Processing sound", "sound", v)
				var accent Accent
				if len(v.Tags) > 0 {
					switch v.Tags[0] {
					case "Canada":
						accent = CA
					case "General-American", "US":
						accent = US
					case "Received-Pronunciation", "UK", "Southern-England":
						accent = UK
					default:
						logger.Warn("Invalid sound tag", "ipa", v.IPA, "tags", v.Tags)
					}
				}
				if v.IPA != "" {
					definition.PhoneticNotations = append(definition.PhoneticNotations, PhoneticNotation{Ipa: v.IPA, Accent: accent})
				}
				if v.MP3URL != "" {
					definition.Sounds = append(definition.Sounds, Sound{Accent: accent, Link: v.MP3URL})
				}
				if v.OGGURL != "" {
					definition.Sounds = append(definition.Sounds, Sound{Accent: accent, Link: v.OGGURL})
				}
			}

			definitions = append(definitions, definition)
		}
	}

	// serializing a set of definition ids as a comma-separated string of numbers
	set := common.NewSet()
	for _, v := range definitions {
		set.Add(v.SourceId)
	}
	cachedValue := strings.Join(set.Values(), ",")
	redis.HSet(context.Background(), "wikitionary", token, cachedValue).Result()

	logger.Debug("finished GetDefinition", "count", len(definitions))
	return definitions, nil
}
