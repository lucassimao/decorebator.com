package nlputils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/xrash/smetrics"
)

type RequestPayload struct {
	Word   string `json:"word"`
	Phrase string `json:"phrase"`
}

type ResponsePayload struct {
	StartIndex int    `json:"start_index"`
	EndIndex   int    `json:"end_index"`
	Error      string `json:"error,omitempty"`
}

func FindTerm(word, phrase string) (int, int, error) {
	url := "http://127.0.0.1:5000/find_term"

	payload := RequestPayload{
		Word:   word,
		Phrase: phrase,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return -1, -1, fmt.Errorf("failed to create request payload %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return -1, -1, fmt.Errorf("failed to create request %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return -1, -1, fmt.Errorf("failed to send request %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("received non-200 status code: %d\n", resp.StatusCode)
		start, end := findClosestMatch(phrase, word)
		return start, end, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return -1, -1, fmt.Errorf("failed to read response body %v", err)
	}

	var responsePayload ResponsePayload
	err = json.Unmarshal(body, &responsePayload)
	if err != nil {
		return -1, -1, fmt.Errorf("failed to unmarshalling response body %v", err)
	}

	if responsePayload.Error != "" {
		return -1, -1, errors.New(responsePayload.Error)
	}

	return responsePayload.StartIndex, responsePayload.EndIndex, nil
}

// FindClosestMatch finds the closest word to the target in the sentence
// and returns the start and end index of the closest matching word.
func findClosestMatch(sentence, target string) (int, int) {
	words := strings.Fields(sentence)
	minDistance := -1
	closestWord := ""
	closestStart := -1

	for _, word := range words {
		distance := smetrics.WagnerFischer(word, target, 1, 1, 2)
		if minDistance == -1 || distance < minDistance {
			minDistance = distance
			closestWord = word
		}
	}

	if closestWord != "" {
		closestStart = strings.Index(sentence, closestWord)
		if closestStart != -1 {
			closestEnd := closestStart + len(closestWord)
			return closestStart, closestEnd
		}
	}

	return -1, -1
}
