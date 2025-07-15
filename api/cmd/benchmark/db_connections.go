package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ConnectionResult holds the result of a single connection attempt
type ConnectionResult struct {
	Success  bool
	Duration time.Duration
	Error    error
}

// TestResult holds the results of testing a specific number of concurrent connections
type TestResult struct {
	Concurrency       int
	SuccessCount      int
	FailureCount      int
	AvgDuration       time.Duration
	MinDuration       time.Duration
	MaxDuration       time.Duration
	ErrorTypes        map[string]int
	TotalTestDuration time.Duration
	FirstError        error // First error for debugging
}

func main() {
	var dbURL string
	var maxConcurrency int
	var timeout int
	var usePool bool
	var maxPoolSize int
	var minPoolSize int

	flag.StringVar(&dbURL, "url", "", "PostgreSQL connection URL (required)")
	flag.IntVar(&maxConcurrency, "max", 1000, "Maximum concurrent connections to test")
	flag.IntVar(&timeout, "timeout", 10, "Connection timeout in seconds")
	flag.BoolVar(&usePool, "pool", false, "Use pgxpool.Pool instead of individual connections")
	flag.IntVar(&maxPoolSize, "max-pool-size", 20, "Maximum pool size (when using -pool)")
	flag.IntVar(&minPoolSize, "min-pool-size", 2, "Minimum pool size (when using -pool)")
	flag.Parse()

	if dbURL == "" {
		log.Fatal("Database URL is required. Use -url flag.")
	}

	fmt.Printf("PostgreSQL Connection Benchmarking Tool - Decorebator Quiz Generation\n")
	fmt.Printf("====================================================================\n")
	fmt.Printf("Database URL: %s\n", maskPassword(dbURL))
	fmt.Printf("Max Concurrency: %d\n", maxConcurrency)
	fmt.Printf("Connection Timeout: %d seconds\n", timeout)
	fmt.Printf("Query Type: Complex Quiz Generation (production-realistic)\n")
	fmt.Printf("Test User: lsimaocosta@gmail.com\n")
	if usePool {
		fmt.Printf("Connection Mode: pgxpool.Pool (MaxPool: %d, MinPool: %d)\n", maxPoolSize, minPoolSize)
	} else {
		fmt.Printf("Connection Mode: Individual pgx.Connect() (no pooling)\n")
	}
	fmt.Printf("\n")

	// Test single connection first
	fmt.Println("Testing single connection...")
	if !testSingleConnection(dbURL, time.Duration(timeout)*time.Second) {
		log.Fatal("Failed to establish single connection. Please check your database URL.")
	}
	fmt.Println("✓ Single connection successful")

	// Get and display the user and wordlist IDs that will be used
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, dbURL)
	if err == nil {
		defer conn.Close(ctx)
		var userID int64
		var wordlistID int64
		if err := conn.QueryRow(ctx, "SELECT id FROM users WHERE email = $1", "lsimaocosta@gmail.com").Scan(&userID); err == nil {
			if err := conn.QueryRow(ctx, "SELECT id FROM wordlists WHERE user_id = $1 LIMIT 1", userID).Scan(&wordlistID); err == nil {
				fmt.Printf("Using User ID: %d, Wordlist ID: %d\n", userID, wordlistID)

				// Show data summary
				var dataCheck struct {
					totalWords      int
					unlearnedWords  int
					withDefinitions int
					withTracking    int
				}

				debugQuery := `
					SELECT 
						COUNT(DISTINCT w.id) as total_words,
						COUNT(DISTINCT CASE WHEN w.learned = FALSE THEN w.id END) as unlearned_words,
						COUNT(DISTINCT CASE WHEN def.meaning IS NOT NULL THEN w.id END) as with_definitions,
						COUNT(DISTINCT CASE WHEN lst.id IS NOT NULL THEN w.id END) as with_tracking
					FROM words w
					LEFT JOIN word_definitions wd ON w.id = wd.word_id
					LEFT JOIN definitions def ON wd.definition_id = def.id
					LEFT JOIN leitner_system_tracking lst ON lst.definition_id = def.id AND lst.user_id = $1
					WHERE w.wordlist_id = $2
				`

				if err := conn.QueryRow(ctx, debugQuery, userID, wordlistID).Scan(
					&dataCheck.totalWords, &dataCheck.unlearnedWords,
					&dataCheck.withDefinitions, &dataCheck.withTracking); err == nil {
					fmt.Printf("Data Summary - Total Words: %d, Unlearned: %d, With Definitions: %d, With Tracking: %d\n",
						dataCheck.totalWords, dataCheck.unlearnedWords, dataCheck.withDefinitions, dataCheck.withTracking)
				}
			}
		}
	}
	fmt.Println()

	// Test progression: 1, 5, 10, 20, 50, 100, 200, 500, 1000
	testLevels := []int{1, 5, 10, 20, 50, 100, 200, 500, 1000}

	// Filter test levels based on maxConcurrency
	var filteredLevels []int
	for _, level := range testLevels {
		if level <= maxConcurrency {
			filteredLevels = append(filteredLevels, level)
		}
	}

	fmt.Printf("Starting concurrent connection tests...\n\n")

	var allResults []TestResult

	for _, concurrency := range filteredLevels {
		fmt.Printf("Testing %d concurrent connections...\n", concurrency)

		// Show memory usage before test
		var memBefore runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&memBefore)

		result := testConcurrentConnections(dbURL, concurrency, time.Duration(timeout)*time.Second, usePool, maxPoolSize, minPoolSize)
		allResults = append(allResults, result)

		// Show memory usage after test
		var memAfter runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&memAfter)

		printTestResult(result)
		fmt.Printf("Memory usage: %.2f MB -> %.2f MB (diff: %.2f MB)\n\n",
			float64(memBefore.Alloc)/1024/1024,
			float64(memAfter.Alloc)/1024/1024,
			float64(memAfter.Alloc-memBefore.Alloc)/1024/1024)

		// If we have significant failures, warn about continuing
		if result.FailureCount > concurrency/2 {
			fmt.Printf("⚠️  Warning: %d/%d connections failed. Database may be reaching limits.\n\n",
				result.FailureCount, concurrency)
		}

		// If all connections failed, stop testing
		if result.SuccessCount == 0 {
			fmt.Printf("❌ All connections failed. Stopping tests.\n\n")
			break
		}
	}

	// Print summary
	printSummary(allResults)
}

// testSingleConnection tests a single connection to verify the database is accessible
func testSingleConnection(dbURL string, timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		fmt.Printf("❌ Connection failed: %v\n", err)
		return false
	}
	defer conn.Close(ctx)

	// Test with a realistic quiz generation query to validate connection
	query := `
		SELECT COUNT(*) as table_count
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name IN ('leitner_system_tracking', 'definitions', 'words', 'wordlists')
	`

	var tableCount int
	err = conn.QueryRow(ctx, query).Scan(&tableCount)
	if err != nil {
		fmt.Printf("❌ Query failed: %v\n", err)
		return false
	}

	// Check if we have the expected tables (should be 4 tables)
	if tableCount < 4 {
		fmt.Printf("❌ Expected database tables missing (found %d, expected 4)\n", tableCount)
		return false
	}

	return true
}

// testConcurrentConnections tests the specified number of concurrent connections
func testConcurrentConnections(dbURL string, concurrency int, timeout time.Duration, usePool bool, maxPoolSize, minPoolSize int) TestResult {
	startTime := time.Now()

	results := make(chan ConnectionResult, concurrency)
	var wg sync.WaitGroup

	// Setup pool if needed
	var pool *pgxpool.Pool
	if usePool {
		// Parse connection string and configure pool
		config, err := pgxpool.ParseConfig(dbURL)
		if err != nil {
			return TestResult{
				Concurrency:       concurrency,
				SuccessCount:      0,
				FailureCount:      concurrency,
				ErrorTypes:        map[string]int{"pool_config_error": concurrency},
				TotalTestDuration: time.Since(startTime),
				FirstError:        fmt.Errorf("failed to parse pool config: %w", err),
			}
		}

		// Configure pool settings (matching production config)
		config.MaxConns = int32(maxPoolSize) //nolint:gosec // G115 - maxPoolSize is controlled by command line flags
		config.MinConns = int32(minPoolSize) //nolint:gosec // G115 - minPoolSize is controlled by command line flags
		config.MaxConnLifetime = time.Hour
		config.MaxConnIdleTime = 15 * time.Minute
		config.ConnConfig.ConnectTimeout = timeout

		pool, err = pgxpool.NewWithConfig(context.Background(), config)
		if err != nil {
			return TestResult{
				Concurrency:       concurrency,
				SuccessCount:      0,
				FailureCount:      concurrency,
				ErrorTypes:        map[string]int{"pool_creation_error": concurrency},
				TotalTestDuration: time.Since(startTime),
				FirstError:        fmt.Errorf("failed to create pool: %w", err),
			}
		}
		defer pool.Close()
	}

	// Launch concurrent connection attempts
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if usePool {
				results <- attemptConnectionWithPool(pool, timeout)
			} else {
				results <- attemptConnection(dbURL, timeout)
			}
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(results)

	// Collect results
	var successCount, failureCount int
	var durations []time.Duration
	errorTypes := make(map[string]int)
	var firstError error // Capture first error for debugging

	for result := range results {
		if result.Success {
			successCount++
			durations = append(durations, result.Duration)
		} else {
			failureCount++
			if firstError == nil {
				firstError = result.Error
			}
			errorType := categorizeError(result.Error)
			errorTypes[errorType]++
		}
	}

	// Calculate duration statistics
	var avgDuration, minDuration, maxDuration time.Duration
	if len(durations) > 0 {
		sort.Slice(durations, func(i, j int) bool {
			return durations[i] < durations[j]
		})

		var total time.Duration
		for _, d := range durations {
			total += d
		}
		avgDuration = total / time.Duration(len(durations))
		minDuration = durations[0]
		maxDuration = durations[len(durations)-1]
	}

	return TestResult{
		Concurrency:       concurrency,
		SuccessCount:      successCount,
		FailureCount:      failureCount,
		AvgDuration:       avgDuration,
		MinDuration:       minDuration,
		MaxDuration:       maxDuration,
		ErrorTypes:        errorTypes,
		TotalTestDuration: time.Since(startTime),
		FirstError:        firstError,
	}
}

// attemptConnection attempts to establish a single connection and run a simple query
func attemptConnection(dbURL string, timeout time.Duration) ConnectionResult {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		return ConnectionResult{
			Success:  false,
			Duration: time.Since(start),
			Error:    err,
		}
	}
	defer conn.Close(ctx)

	// Test the connection with a realistic quiz generation query
	// This is the actual complex query used in production for quiz generation
	query := `
		WITH definition_priorities AS (
			SELECT 
				def.id, lst.id AS lst_id, def.token, 
				def.part_of_speech, def.language, def.is_verb_type, def.meaning, def.examples, 
				def.inflections, lst.box_id, def.sounds, def.phonetic_notations, 
				di.url as image_url, di.description as image_description, 
				wd.word_id AS word_id,
				lst.updated_at,
				-- Calculate hours since last review (NULL = never reviewed = maximum hours)
				CASE 
					WHEN lst.updated_at IS NULL THEN 999999  -- New words get maximum hours
					ELSE EXTRACT(EPOCH FROM (NOW() - lst.updated_at))/3600
				END as hours_since_review,
				-- Get target interval hours based on box
				CASE lst.box_id
					WHEN 1 THEN 0     -- Always available
					WHEN 2 THEN 6     -- 6 hours
					WHEN 3 THEN 24    -- 1 day
					WHEN 4 THEN 72    -- 3 days
					WHEN 5 THEN 168   -- 1 week
					WHEN 6 THEN 336   -- 2 weeks
					WHEN 7 THEN 720   -- 1 month
				END as target_hours,
				-- Calculate progress ratio (how close to target interval)
				CASE 
					WHEN lst.box_id = 1 THEN 1.0  -- Box 1 always 100% ready
					WHEN lst.updated_at IS NULL THEN 1.0  -- New words always ready
					ELSE (CASE 
						WHEN lst.updated_at IS NULL THEN 999999  
						ELSE EXTRACT(EPOCH FROM (NOW() - lst.updated_at))/3600
					END) / NULLIF(CASE lst.box_id
						WHEN 2 THEN 6
						WHEN 3 THEN 24
						WHEN 4 THEN 72
						WHEN 5 THEN 168
						WHEN 6 THEN 336
						WHEN 7 THEN 720
					END, 0)
				END as progress_ratio
			FROM leitner_system_tracking lst 
			JOIN definitions def ON lst.definition_id = def.id
			JOIN word_definitions wd ON def.id = wd.definition_id
			JOIN words w ON wd.word_id = w.id
			LEFT JOIN definition_images di ON di.definition_id = def.id AND di.is_visible=TRUE
			WHERE 
				lst.user_id = $1
				AND w.wordlist_id = $2
				AND w.learned = FALSE
				AND def.meaning IS NOT NULL
				-- Exclude temporarily skipped definitions
				AND (lst.temporarily_skipped_until IS NULL OR lst.temporarily_skipped_until < NOW())
		),
		weighted_definitions AS (
			SELECT 
				id, token, part_of_speech, language, is_verb_type, meaning, examples, inflections, 
				lst_id, box_id, sounds, phonetic_notations, 
				COALESCE(image_url,'') as image_url, word_id, COALESCE(image_description,'') as image_description,
				hours_since_review, progress_ratio, updated_at,
				-- Pure priority based on progress ratio - no complex calculations
				CASE 
					WHEN progress_ratio >= 1.0 THEN 1000    -- Overdue (highest priority)
					WHEN progress_ratio >= 0.8 THEN 800     -- Due soon
					WHEN progress_ratio >= 0.5 THEN 500     -- Available
					ELSE FLOOR(progress_ratio * 100)        -- Early (proportional priority)
				END as selection_weight
			FROM definition_priorities
		),
		selected_definition AS (
			SELECT 
				id, token, part_of_speech, language, is_verb_type, meaning, examples, inflections, 
				lst_id, box_id, sounds, phonetic_notations, 
				image_url, word_id, image_description, hours_since_review, progress_ratio, updated_at
			FROM weighted_definitions
			-- Pure priority-based selection: respects spaced repetition science exactly
			-- No artificial randomness - selection is 100% predictable and deterministic
			ORDER BY 
				-- Primary sort: Highest weight wins (respects spaced repetition intervals)
				selection_weight DESC,
				-- Secondary sort: Among equal weights, prioritize oldest reviewed (new words = NULL = oldest)
				COALESCE(updated_at, '1970-01-01') ASC,
				-- Final tiebreaker: Lowest definition ID wins (deterministic)
				id ASC
			LIMIT 1
		)
		SELECT 
			sd.id, sd.token, sd.part_of_speech, sd.language, sd.is_verb_type, sd.meaning, sd.examples, sd.inflections, 
			sd.lst_id, sd.box_id, sd.sounds, sd.phonetic_notations, 
			sd.image_url, sd.word_id, sd.image_description,
			sd.hours_since_review, sd.progress_ratio,
			-- Aggregate example audio files into JSON array
			COALESCE(
				JSON_AGG(
					JSON_BUILD_OBJECT(
						'id', dea.id,
						'definitionId', dea.definition_id,
						'exampleText', dea.example_text,
						'exampleHash', dea.example_hash,
						'audioUrl', dea.audio_url,
						'inflectionType', COALESCE(dea.inflection_type, ''),
						'createdAt', dea.created_at
					) ORDER BY dea.created_at DESC
				) FILTER (WHERE dea.id IS NOT NULL),
				'[]'::json
			) as example_audio_files
		FROM selected_definition sd
		LEFT JOIN definition_example_audio dea ON dea.definition_id = sd.id
		GROUP BY sd.id, sd.token, sd.part_of_speech, sd.language, sd.is_verb_type, sd.meaning, sd.examples, sd.inflections, 
				 sd.lst_id, sd.box_id, sd.sounds, sd.phonetic_notations, 
				 sd.image_url, sd.word_id, sd.image_description, sd.hours_since_review, sd.progress_ratio, sd.updated_at;
	`

	// First, get the actual user ID for lsimaocosta@gmail.com
	var userID int64
	err = conn.QueryRow(ctx, "SELECT id FROM users WHERE email = $1", "lsimaocosta@gmail.com").Scan(&userID)
	if err != nil {
		return ConnectionResult{
			Success:  false,
			Duration: time.Since(start),
			Error:    fmt.Errorf("failed to find user lsimaocosta@gmail.com: %w", err),
		}
	}

	// Get one of their wordlists
	var wordlistID int64
	err = conn.QueryRow(ctx, "SELECT id FROM wordlists WHERE user_id = $1 LIMIT 1 ORDER BY RANDOM()", userID).Scan(&wordlistID)
	if err != nil {
		return ConnectionResult{
			Success:  false,
			Duration: time.Since(start),
			Error:    fmt.Errorf("failed to find wordlist for user %d: %w", userID, err),
		}
	}

	// Debug: Check if user has any data in the leitner system
	var dataCheck struct {
		totalWords      int
		unlearnedWords  int
		withDefinitions int
		withTracking    int
	}

	debugQuery := `
		SELECT 
			COUNT(DISTINCT w.id) as total_words,
			COUNT(DISTINCT CASE WHEN w.learned = FALSE THEN w.id END) as unlearned_words,
			COUNT(DISTINCT CASE WHEN def.meaning IS NOT NULL THEN w.id END) as with_definitions,
			COUNT(DISTINCT CASE WHEN lst.id IS NOT NULL THEN w.id END) as with_tracking
		FROM words w
		LEFT JOIN word_definitions wd ON w.id = wd.word_id
		LEFT JOIN definitions def ON wd.definition_id = def.id
		LEFT JOIN leitner_system_tracking lst ON lst.definition_id = def.id AND lst.user_id = $1
		WHERE w.wordlist_id = $2
	`

	err = conn.QueryRow(ctx, debugQuery, userID, wordlistID).Scan(
		&dataCheck.totalWords, &dataCheck.unlearnedWords,
		&dataCheck.withDefinitions, &dataCheck.withTracking)
	if err != nil {
		return ConnectionResult{
			Success:  false,
			Duration: time.Since(start),
			Error:    fmt.Errorf("debug query failed for user %d, wordlist %d: %w", userID, wordlistID, err),
		}
	}

	// If no tracking data, let's use a simpler query that just checks basic data
	if dataCheck.withTracking == 0 {
		// Fallback to a simpler query that just gets word data
		simpleQuery := `
			SELECT w.id, w.name, w.wordlist_id, w.learned, 
				   COALESCE(def.meaning, 'No meaning') as meaning,
				   COALESCE(def.part_of_speech, 'unknown') as part_of_speech
			FROM words w
			LEFT JOIN word_definitions wd ON w.id = wd.word_id
			LEFT JOIN definitions def ON wd.definition_id = def.id
			WHERE w.wordlist_id = $1 AND w.user_id = $2
			LIMIT 1
		`

		rows, queryErr := conn.Query(ctx, simpleQuery, wordlistID, userID)
		if queryErr != nil {
			return ConnectionResult{
				Success:  false,
				Duration: time.Since(start),
				Error:    fmt.Errorf("simple query failed (no tracking data) for user %d, wordlist %d: %w", userID, wordlistID, queryErr),
			}
		}
		defer rows.Close()

		var resultCount int
		for rows.Next() {
			var id, wordlistIDLocal int64
			var name, meaning, partOfSpeech string
			var learned bool

			scanErr := rows.Scan(&id, &name, &wordlistIDLocal, &learned, &meaning, &partOfSpeech)
			if scanErr != nil {
				return ConnectionResult{
					Success:  false,
					Duration: time.Since(start),
					Error:    fmt.Errorf("scan failed in simple query: %w", scanErr),
				}
			}
			resultCount++
		}

		// Check for any errors during iteration
		if rowsErr := rows.Err(); rowsErr != nil {
			return ConnectionResult{
				Success:  false,
				Duration: time.Since(start),
				Error:    fmt.Errorf("iteration error in simple query: %w", rowsErr),
			}
		}

		return ConnectionResult{
			Success:  true,
			Duration: time.Since(start),
			Error:    nil,
		}
	}

	// Use the actual user and wordlist IDs with the complex query
	rows, err := conn.Query(ctx, query, userID, wordlistID)
	if err != nil {
		return ConnectionResult{
			Success:  false,
			Duration: time.Since(start),
			Error:    fmt.Errorf("quiz query failed for user %d, wordlist %d (data: %+v): %w", userID, wordlistID, dataCheck, err),
		}
	}
	defer rows.Close()

	// Process the results to make it realistic
	var resultCount int
	for rows.Next() {
		var (
			id, lstID, boxID, wordID        int64
			token, partOfSpeech, language   string
			isVerbType                      bool
			meaning                         *string
			examples, inflections           interface{} // These are arrays in PostgreSQL
			sounds, phoneticNotations       *string
			imageURL, imageDescription      string
			hoursSinceReview, progressRatio float64
			exampleAudioFiles               string // Change from []byte to string
		)

		err = rows.Scan(&id, &token, &partOfSpeech, &language, &isVerbType,
			&meaning, &examples, &inflections, &lstID, &boxID, &sounds,
			&phoneticNotations, &imageURL, &wordID, &imageDescription,
			&hoursSinceReview, &progressRatio, &exampleAudioFiles)
		if err != nil {
			return ConnectionResult{
				Success:  false,
				Duration: time.Since(start),
				Error:    fmt.Errorf("scan failed in complex query: %w", err),
			}
		}
		resultCount++
	}

	// Check for any errors during iteration
	if err = rows.Err(); err != nil {
		return ConnectionResult{
			Success:  false,
			Duration: time.Since(start),
			Error:    fmt.Errorf("iteration error in complex query: %w", err),
		}
	}

	return ConnectionResult{
		Success:  true,
		Duration: time.Since(start),
		Error:    nil,
	}
}

// attemptConnectionWithPool attempts to get a connection from the pool and run the complex query
func attemptConnectionWithPool(pool *pgxpool.Pool, timeout time.Duration) ConnectionResult {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Get connection from pool
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return ConnectionResult{
			Success:  false,
			Duration: time.Since(start),
			Error:    fmt.Errorf("failed to acquire connection from pool: %w", err),
		}
	}
	defer conn.Release()

	// First, get the actual user ID for lsimaocosta@gmail.com
	var userID int64
	err = conn.QueryRow(ctx, "SELECT id FROM users WHERE email = $1", "lsimaocosta@gmail.com").Scan(&userID)
	if err != nil {
		return ConnectionResult{
			Success:  false,
			Duration: time.Since(start),
			Error:    fmt.Errorf("failed to find user lsimaocosta@gmail.com: %w", err),
		}
	}

	// Get one of their wordlists
	var wordlistID int64
	err = conn.QueryRow(ctx, "SELECT id FROM wordlists WHERE user_id = $1 LIMIT 1", userID).Scan(&wordlistID)
	if err != nil {
		return ConnectionResult{
			Success:  false,
			Duration: time.Since(start),
			Error:    fmt.Errorf("failed to find wordlist for user %d: %w", userID, err),
		}
	}

	// Debug: Check if user has any data in the leitner system
	var dataCheck struct {
		totalWords      int
		unlearnedWords  int
		withDefinitions int
		withTracking    int
	}

	debugQuery := `
		SELECT 
			COUNT(DISTINCT w.id) as total_words,
			COUNT(DISTINCT CASE WHEN w.learned = FALSE THEN w.id END) as unlearned_words,
			COUNT(DISTINCT CASE WHEN def.meaning IS NOT NULL THEN w.id END) as with_definitions,
			COUNT(DISTINCT CASE WHEN lst.id IS NOT NULL THEN w.id END) as with_tracking
		FROM words w
		LEFT JOIN word_definitions wd ON w.id = wd.word_id
		LEFT JOIN definitions def ON wd.definition_id = def.id
		LEFT JOIN leitner_system_tracking lst ON lst.definition_id = def.id AND lst.user_id = $1
		WHERE w.wordlist_id = $2
	`

	err = conn.QueryRow(ctx, debugQuery, userID, wordlistID).Scan(
		&dataCheck.totalWords, &dataCheck.unlearnedWords,
		&dataCheck.withDefinitions, &dataCheck.withTracking)
	if err != nil {
		return ConnectionResult{
			Success:  false,
			Duration: time.Since(start),
			Error:    fmt.Errorf("debug query failed for user %d, wordlist %d: %w", userID, wordlistID, err),
		}
	}

	// If no tracking data, let's use a simpler query that just checks basic data
	if dataCheck.withTracking == 0 {
		// Fallback to a simpler query that just gets word data
		simpleQuery := `
			SELECT w.id, w.name, w.wordlist_id, w.learned, 
				   COALESCE(def.meaning, 'No meaning') as meaning,
				   COALESCE(def.part_of_speech, 'unknown') as part_of_speech
			FROM words w
			LEFT JOIN word_definitions wd ON w.id = wd.word_id
			LEFT JOIN definitions def ON wd.definition_id = def.id
			WHERE w.wordlist_id = $1 AND w.user_id = $2
			LIMIT 1
		`

		rows, queryErr := conn.Query(ctx, simpleQuery, wordlistID, userID)
		if queryErr != nil {
			return ConnectionResult{
				Success:  false,
				Duration: time.Since(start),
				Error:    fmt.Errorf("simple query failed (no tracking data) for user %d, wordlist %d: %w", userID, wordlistID, queryErr),
			}
		}
		defer rows.Close()

		var resultCount int
		for rows.Next() {
			var id, wordlistIDLocal int64
			var name, meaning, partOfSpeech string
			var learned bool

			scanErr := rows.Scan(&id, &name, &wordlistIDLocal, &learned, &meaning, &partOfSpeech)
			if scanErr != nil {
				return ConnectionResult{
					Success:  false,
					Duration: time.Since(start),
					Error:    fmt.Errorf("scan failed in simple query: %w", scanErr),
				}
			}
			resultCount++
		}

		// Check for any errors during iteration
		if rowsErr := rows.Err(); rowsErr != nil {
			return ConnectionResult{
				Success:  false,
				Duration: time.Since(start),
				Error:    fmt.Errorf("iteration error in simple query: %w", rowsErr),
			}
		}

		return ConnectionResult{
			Success:  true,
			Duration: time.Since(start),
			Error:    nil,
		}
	}

	// Test the connection with the realistic quiz generation query
	query := `
		WITH definition_priorities AS (
			SELECT 
				def.id, lst.id AS lst_id, def.token, 
				def.part_of_speech, def.language, def.is_verb_type, def.meaning, def.examples, 
				def.inflections, lst.box_id, def.sounds, def.phonetic_notations, 
				di.url as image_url, di.description as image_description, 
				wd.word_id AS word_id,
				lst.updated_at,
				-- Calculate hours since last review (NULL = never reviewed = maximum hours)
				CASE 
					WHEN lst.updated_at IS NULL THEN 999999  -- New words get maximum hours
					ELSE EXTRACT(EPOCH FROM (NOW() - lst.updated_at))/3600
				END as hours_since_review,
				-- Get target interval hours based on box
				CASE lst.box_id
					WHEN 1 THEN 0     -- Always available
					WHEN 2 THEN 6     -- 6 hours
					WHEN 3 THEN 24    -- 1 day
					WHEN 4 THEN 72    -- 3 days
					WHEN 5 THEN 168   -- 1 week
					WHEN 6 THEN 336   -- 2 weeks
					WHEN 7 THEN 720   -- 1 month
				END as target_hours,
				-- Calculate progress ratio (how close to target interval)
				CASE 
					WHEN lst.box_id = 1 THEN 1.0  -- Box 1 always 100% ready
					WHEN lst.updated_at IS NULL THEN 1.0  -- New words always ready
					ELSE (CASE 
						WHEN lst.updated_at IS NULL THEN 999999  
						ELSE EXTRACT(EPOCH FROM (NOW() - lst.updated_at))/3600
					END) / NULLIF(CASE lst.box_id
						WHEN 2 THEN 6
						WHEN 3 THEN 24
						WHEN 4 THEN 72
						WHEN 5 THEN 168
						WHEN 6 THEN 336
						WHEN 7 THEN 720
					END, 0)
				END as progress_ratio
			FROM leitner_system_tracking lst 
			JOIN definitions def ON lst.definition_id = def.id
			JOIN word_definitions wd ON def.id = wd.definition_id
			JOIN words w ON wd.word_id = w.id
			LEFT JOIN definition_images di ON di.definition_id = def.id AND di.is_visible=TRUE
			WHERE 
				lst.user_id = $1
				AND w.wordlist_id = $2
				AND w.learned = FALSE
				AND def.meaning IS NOT NULL
				-- Exclude temporarily skipped definitions
				AND (lst.temporarily_skipped_until IS NULL OR lst.temporarily_skipped_until < NOW())
		),
		weighted_definitions AS (
			SELECT 
				id, token, part_of_speech, language, is_verb_type, meaning, examples, inflections, 
				lst_id, box_id, sounds, phonetic_notations, 
				COALESCE(image_url,'') as image_url, word_id, COALESCE(image_description,'') as image_description,
				hours_since_review, progress_ratio, updated_at,
				-- Pure priority based on progress ratio - no complex calculations
				CASE 
					WHEN progress_ratio >= 1.0 THEN 1000    -- Overdue (highest priority)
					WHEN progress_ratio >= 0.8 THEN 800     -- Due soon
					WHEN progress_ratio >= 0.5 THEN 500     -- Available
					ELSE FLOOR(progress_ratio * 100)        -- Early (proportional priority)
				END as selection_weight
			FROM definition_priorities
		),
		selected_definition AS (
			SELECT 
				id, token, part_of_speech, language, is_verb_type, meaning, examples, inflections, 
				lst_id, box_id, sounds, phonetic_notations, 
				image_url, word_id, image_description, hours_since_review, progress_ratio, updated_at
			FROM weighted_definitions
			-- Pure priority-based selection: respects spaced repetition science exactly
			-- No artificial randomness - selection is 100% predictable and deterministic
			ORDER BY 
				-- Primary sort: Highest weight wins (respects spaced repetition intervals)
				selection_weight DESC,
				-- Secondary sort: Among equal weights, prioritize oldest reviewed (new words = NULL = oldest)
				COALESCE(updated_at, '1970-01-01') ASC,
				-- Final tiebreaker: Lowest definition ID wins (deterministic)
				id ASC
			LIMIT 1
		)
		SELECT 
			sd.id, sd.token, sd.part_of_speech, sd.language, sd.is_verb_type, sd.meaning, sd.examples, sd.inflections, 
			sd.lst_id, sd.box_id, sd.sounds, sd.phonetic_notations, 
			sd.image_url, sd.word_id, sd.image_description,
			sd.hours_since_review, sd.progress_ratio,
			-- Aggregate example audio files into JSON array
			COALESCE(
				JSON_AGG(
					JSON_BUILD_OBJECT(
						'id', dea.id,
						'definitionId', dea.definition_id,
						'exampleText', dea.example_text,
						'exampleHash', dea.example_hash,
						'audioUrl', dea.audio_url,
						'inflectionType', COALESCE(dea.inflection_type, ''),
						'createdAt', dea.created_at
					) ORDER BY dea.created_at DESC
				) FILTER (WHERE dea.id IS NOT NULL),
				'[]'::json
			) as example_audio_files
		FROM selected_definition sd
		LEFT JOIN definition_example_audio dea ON dea.definition_id = sd.id
		GROUP BY sd.id, sd.token, sd.part_of_speech, sd.language, sd.is_verb_type, sd.meaning, sd.examples, sd.inflections, 
				 sd.lst_id, sd.box_id, sd.sounds, sd.phonetic_notations, 
				 sd.image_url, sd.word_id, sd.image_description, sd.hours_since_review, sd.progress_ratio, sd.updated_at;
	`

	// Use the actual user and wordlist IDs with the complex query
	rows, err := conn.Query(ctx, query, userID, wordlistID)
	if err != nil {
		return ConnectionResult{
			Success:  false,
			Duration: time.Since(start),
			Error:    fmt.Errorf("quiz query failed for user %d, wordlist %d (data: %+v): %w", userID, wordlistID, dataCheck, err),
		}
	}
	defer rows.Close()

	// Process the results to make it realistic
	var resultCount int
	for rows.Next() {
		var (
			id, lstID, boxID, wordID        int64
			token, partOfSpeech, language   string
			isVerbType                      bool
			meaning                         *string
			examples, inflections           interface{} // These are arrays in PostgreSQL
			sounds, phoneticNotations       *string
			imageURL, imageDescription      string
			hoursSinceReview, progressRatio float64
			exampleAudioFiles               string // Change from []byte to string
		)

		err = rows.Scan(&id, &token, &partOfSpeech, &language, &isVerbType,
			&meaning, &examples, &inflections, &lstID, &boxID, &sounds,
			&phoneticNotations, &imageURL, &wordID, &imageDescription,
			&hoursSinceReview, &progressRatio, &exampleAudioFiles)
		if err != nil {
			return ConnectionResult{
				Success:  false,
				Duration: time.Since(start),
				Error:    fmt.Errorf("scan failed in complex query: %w", err),
			}
		}
		resultCount++
	}

	// Check for any errors during iteration
	if err = rows.Err(); err != nil {
		return ConnectionResult{
			Success:  false,
			Duration: time.Since(start),
			Error:    fmt.Errorf("iteration error in complex query: %w", err),
		}
	}

	return ConnectionResult{
		Success:  true,
		Duration: time.Since(start),
		Error:    nil,
	}
}

// categorizeError categorizes database connection errors
func categorizeError(err error) string {
	if err == nil {
		return "unknown"
	}

	errStr := err.Error()

	// Common PostgreSQL connection errors
	if contains(errStr, "connection refused") {
		return "connection_refused"
	}
	if contains(errStr, "timeout") || contains(errStr, "deadline exceeded") {
		return "timeout"
	}
	if contains(errStr, "too many connections") {
		return "too_many_connections"
	}
	if contains(errStr, "authentication failed") {
		return "auth_failed"
	}
	if contains(errStr, "database") && contains(errStr, "does not exist") {
		return "database_not_found"
	}
	if contains(errStr, "no such host") {
		return "host_not_found"
	}
	if contains(errStr, "column") && contains(errStr, "does not exist") {
		return "column_not_found"
	}
	if contains(errStr, "relation") && contains(errStr, "does not exist") {
		return "table_not_found"
	}
	if contains(errStr, "syntax error") {
		return "syntax_error"
	}

	return "other"
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
				containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// printTestResult prints the results of a single test
func printTestResult(result TestResult) {
	if result.SuccessCount == result.Concurrency {
		fmt.Printf("✓ Success: %d/%d connections (avg: %v, min: %v, max: %v)\n",
			result.SuccessCount, result.Concurrency, result.AvgDuration, result.MinDuration, result.MaxDuration)
	} else {
		fmt.Printf("⚠️  Partial: %d/%d connections (avg: %v, min: %v, max: %v)\n",
			result.SuccessCount, result.Concurrency, result.AvgDuration, result.MinDuration, result.MaxDuration)

		if len(result.ErrorTypes) > 0 {
			fmt.Printf("   Errors: ")
			first := true
			for errorType, count := range result.ErrorTypes {
				if !first {
					fmt.Printf(", ")
				}
				fmt.Printf("%s (%d)", errorType, count)
				first = false
			}
			fmt.Printf("\n")

			// Show first error for debugging
			if result.FirstError != nil {
				fmt.Printf("   First error: %v\n", result.FirstError)
			}
		}
	}

	fmt.Printf("   Test duration: %v\n", result.TotalTestDuration)
}

// printSummary prints a summary of all test results
func printSummary(results []TestResult) {
	fmt.Printf("Summary\n")
	fmt.Printf("=======\n")

	maxSuccessful := 0
	for _, result := range results {
		if result.SuccessCount == result.Concurrency {
			maxSuccessful = result.Concurrency
		}
	}

	fmt.Printf("Maximum successful concurrent connections: %d\n", maxSuccessful)

	// Find the breaking point
	for _, result := range results {
		if result.FailureCount > 0 {
			fmt.Printf("First failure at: %d concurrent connections (%d succeeded, %d failed)\n",
				result.Concurrency, result.SuccessCount, result.FailureCount)
			break
		}
	}

	// Performance summary
	fmt.Printf("\nPerformance Summary:\n")
	fmt.Printf("Concurrency | Success | Avg Duration | Total Duration\n")
	fmt.Printf("------------|---------|--------------|---------------\n")
	for _, result := range results {
		fmt.Printf("%11d | %7d | %12v | %14v\n",
			result.Concurrency, result.SuccessCount, result.AvgDuration, result.TotalTestDuration)
	}
}

// maskPassword masks the password in a database URL for display
func maskPassword(dbURL string) string {
	// Simple password masking - find pattern between :// and @
	start := 0
	for i := 0; i < len(dbURL)-2; i++ {
		if dbURL[i:i+3] == "://" {
			start = i + 3
			break
		}
	}

	if start == 0 {
		return dbURL
	}

	// Find the @ symbol
	atPos := -1
	for i := start; i < len(dbURL); i++ {
		if dbURL[i] == '@' {
			atPos = i
			break
		}
	}

	if atPos == -1 {
		return dbURL
	}

	// Find the : before password
	colonPos := -1
	for i := start; i < atPos; i++ {
		if dbURL[i] == ':' {
			colonPos = i
			break
		}
	}

	if colonPos == -1 {
		return dbURL
	}

	// Replace password with asterisks
	return dbURL[:colonPos+1] + "****" + dbURL[atPos:]
}
