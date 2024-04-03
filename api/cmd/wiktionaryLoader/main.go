// Load word definitions from Wiktionary and save them to the database.
// Input file is obtained from https://github.com/tatuylonen/wiktextract
// https://kaikki.org/dictionary/rawdata.html
// monitor goroutines with: pgrep -f exe/wiktionaryLoader | xargs ps -o nlwp | awk 'NR>1' | awk '{sum += $1} END {print sum}'
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"

	"decorebator.com/internal/common"
	_ "github.com/joho/godotenv/autoload"
)

type WiktionaryData struct {
	Word         string `json:"word"`
	LanguageCode string `json:"lang_code"`
}

var db, err = common.GetDBConnection()
var wg sync.WaitGroup

// semaphore to limit the number of concurrent goroutines that run the saveBatch function
var sem = make(chan bool, runtime.NumCPU())

func init() {
	fmt.Printf("Number of workers: %d\n", runtime.NumCPU())
	if err != nil {
		log.Fatalf("Failed to open db: %v", err)
	}
}

func main() {
	defer db.Close()

	if len(os.Args) < 2 {
		log.Fatalln("Please provide the path to the input file")
	}

	filePath := os.Args[1]

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	maxCapacity := 20 * 1024 * 1024 // 10MB buffer to handle large lines. Default is 64kb
	buf := make([]byte, 0, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	var BATCH_SIZE = 1000
	var jsonData = make([]string, 0, BATCH_SIZE)

	for scanner.Scan() {
		line := scanner.Text()

		if len(jsonData) == BATCH_SIZE {
			wg.Add(1)
			var clonedData = make([]string, len(jsonData))
			// acquire semaphore before starting a new goroutine
			sem <- true
			copy(clonedData, jsonData)
			go saveBatch(clonedData)
			jsonData = jsonData[:0]
		} else {
			jsonData = append(jsonData, line)
		}
	}

	// remaining data
	if len(jsonData) > 0 && len(jsonData) < BATCH_SIZE {
		go saveBatch(jsonData)
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	wg.Wait()
	log.Println("Finished")
}

func saveBatch(jsonData []string) {
	defer wg.Done()
	defer func() { <-sem }() // release semaphore

	var positionalParams = make([]any, 0, 2*len(jsonData)) // ( word, json, word_2, json_2 ...)

	var stringBuilder strings.Builder
	stringBuilder.WriteString(`INSERT INTO wiktionary (word, data) VALUES `)

	// Counter for positional parameters
	counter := 0

	for _, data := range jsonData {
		var jsonObj WiktionaryData
		if err := json.Unmarshal([]byte(data), &jsonObj); err == nil {
			if jsonObj.LanguageCode != "en" || len(jsonObj.Word) == 0 {
				continue
			}
			positionalParams = append(positionalParams, jsonObj.Word, data)
			stringBuilder.WriteString(fmt.Sprintf("($%d, $%d),", counter+1, counter+2))
			counter += 2
		} else {
			log.Fatalf("Failed to unmarshal json data: %v %v", err, data)
		}
	}

	if len(positionalParams) == 0 {
		return
	}

	var sql = stringBuilder.String() // remove the trailing comma
	var sqlNoTrailingComma = sql[:len(sql)-1]

	result, err := db.Exec(context.Background(), sqlNoTrailingComma, positionalParams...)
	if err != nil {
		fmt.Println(sql)
		fmt.Println(len(positionalParams))
		log.Fatalf("Failed to insert data: %v\n ", err)
	}

	fmt.Printf("Inserted %d rows\n", result.RowsAffected())
}
