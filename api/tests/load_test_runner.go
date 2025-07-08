package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"

	httpHandler "decorebator.com/internal/http"
	"decorebator.com/internal/model"
	"github.com/brianvoe/gofakeit/v7"
)

// Test configuration
type Config struct {
	BaseURL       string
	Users         int
	Duration      time.Duration
	ClientTimeout time.Duration
}

// Request metrics
type RequestMetric struct {
	Endpoint     string
	Duration     time.Duration
	Success      bool
	ErrorMessage string
	StatusCode   int
}

// User data for testing
type User struct {
	FirstName string `json:"FirstName"`
	LastName  string `json:"LastName"`
	Email     string `json:"Email"`
	Password  string `json:"Password"`
}

type LoginRequest struct {
	Email    string `json:"Email"`
	Password string `json:"Password"`
}

type WordlistRequest struct {
	Name         string `json:"Name"`
	LanguageCode string `json:"LanguageCode"`
}

type WordRequest struct {
	Name string `json:"Name"`
}

// Reusing existing types from the codebase

// API response types
type WordlistResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Using existing Quiz type from model package

// Predefined words for consistent testing
var testWords = []string{
	"stenciling",
	"furled",
	"klaxons",
	"scuffed",
	"suckling",
	"quaint",
	"counteract",
	"pheasant",
	"piled",
	"gaunt",
	"voiture",
	"sod",
	"sneers",
	"panoply",
	"affordances",
	"parquet",
	"pastureland",
	"liver",
	"slant",
	"plump",
	"barracks",
	"mingle",
	"gills",
	"livre",
	"mating",
	"sighing",
	"lurched",
	"stern",
	"surge",
	"steeply",
	"languidly",
	"whopping",
	"parade",
	"louche",
	"pockmarked",
	"jutting",
	"groggy",
	"bargemen",
	"stringing",
	"sash",
	"sunken",
	"skiff",
	"leaders",
	"lance",
	"yell",
	"spring",
	"frightfully",
	"skinned",
	"mass",
	"simpering",
	"harness",
	"trotted",
	"conceit",
	"gentle",
	"clattering",
	"treachery",
	"fin",
	"sail",
	"mushy",
	"gaffed",
	"stirring",
	"cove",
	"wickets",
	"concourse",
	"defray",
	"merriment",
}

// Metrics collector
type MetricsCollector struct {
	mu      sync.Mutex
	metrics []RequestMetric
}

func (mc *MetricsCollector) Add(metric RequestMetric) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.metrics = append(mc.metrics, metric)
}

func (mc *MetricsCollector) GetStats() map[string]EndpointStats {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	stats := make(map[string]EndpointStats)

	for _, metric := range mc.metrics {
		if _, exists := stats[metric.Endpoint]; !exists {
			stats[metric.Endpoint] = EndpointStats{
				Endpoint:  metric.Endpoint,
				Durations: []time.Duration{},
			}
		}

		s := stats[metric.Endpoint]
		s.Total++
		s.Durations = append(s.Durations, metric.Duration)

		if metric.Success {
			s.Success++
		} else {
			s.Errors++
			if s.ErrorMessages == nil {
				s.ErrorMessages = make(map[string]int)
			}
			s.ErrorMessages[metric.ErrorMessage]++
		}

		stats[metric.Endpoint] = s
	}

	// Calculate percentiles for each endpoint
	for endpoint, stat := range stats {
		if len(stat.Durations) > 0 {
			sort.Slice(stat.Durations, func(i, j int) bool {
				return stat.Durations[i] < stat.Durations[j]
			})

			stat.AvgDuration = calculateAverage(stat.Durations)
			stat.P95Duration = calculatePercentile(stat.Durations, 0.95)
			stat.P99Duration = calculatePercentile(stat.Durations, 0.99)

			stats[endpoint] = stat
		}
	}

	return stats
}

type EndpointStats struct {
	Endpoint      string
	Total         int
	Success       int
	Errors        int
	ErrorMessages map[string]int
	Durations     []time.Duration
	AvgDuration   time.Duration
	P95Duration   time.Duration
	P99Duration   time.Duration
}

// Performance thresholds based on 2025 industry standards
type PerformanceThresholds struct {
	Excellent time.Duration // < 100ms - Perceived as instantaneous
	Good      time.Duration // < 500ms - Good user experience  
	Acceptable time.Duration // < 1000ms - Acceptable for most apps
	Poor      time.Duration // < 2000ms - May impact user experience
	// > 2000ms is considered unacceptable
}

var thresholds = PerformanceThresholds{
	Excellent:  100 * time.Millisecond,
	Good:       500 * time.Millisecond,
	Acceptable: 1000 * time.Millisecond,
	Poor:       2000 * time.Millisecond,
}

func getPerformanceRating(duration time.Duration) (string, string) {
	switch {
	case duration <= thresholds.Excellent:
		return "🟢 EXCELLENT", "Perceived as instantaneous"
	case duration <= thresholds.Good:
		return "🟢 GOOD", "Good user experience"
	case duration <= thresholds.Acceptable:
		return "🟡 ACCEPTABLE", "Acceptable for most apps"
	case duration <= thresholds.Poor:
		return "🟠 POOR", "May impact user experience"
	default:
		return "🔴 UNACCEPTABLE", "Users likely to abandon"
	}
}

func calculateAverage(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	var total time.Duration
	for _, d := range durations {
		total += d
	}
	return total / time.Duration(len(durations))
}

func calculatePercentile(durations []time.Duration, percentile float64) time.Duration {
	if len(durations) == 0 {
		return 0
	}

	index := int(float64(len(durations)) * percentile)
	if index >= len(durations) {
		index = len(durations) - 1
	}

	return durations[index]
}

// HTTP client wrapper for metrics collection
type MetricsHTTPClient struct {
	client    *http.Client
	collector *MetricsCollector
}

func (m *MetricsHTTPClient) Do(req *http.Request, endpoint string) (*http.Response, error) {
	start := time.Now()

	resp, err := m.client.Do(req)

	duration := time.Since(start)
	success := err == nil && resp != nil && resp.StatusCode < 400

	errorMessage := ""
	statusCode := 0
	if err != nil {
		errorMessage = err.Error()
	} else if resp != nil {
		statusCode = resp.StatusCode
		if resp.StatusCode >= 400 {
			errorMessage = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
	}

	m.collector.Add(RequestMetric{
		Endpoint:     endpoint,
		Duration:     duration,
		Success:      success,
		ErrorMessage: errorMessage,
		StatusCode:   statusCode,
	})

	return resp, err
}

// User simulator
type UserSimulator struct {
	client    *MetricsHTTPClient
	config    Config
	userID    int
	authToken string
}

func NewUserSimulator(client *MetricsHTTPClient, config Config, userID int) *UserSimulator {
	return &UserSimulator{
		client: client,
		config: config,
		userID: userID,
	}
}

func (us *UserSimulator) SimulateUserJourney() error {
	var userCreated bool
	var journeyErr error

	// Ensure cleanup always happens if user was created, even if journey fails
	defer func() {
		if userCreated && us.authToken != "" {
			if err := us.deleteUser(); err != nil {
				log.Printf("User %d cleanup failed: %v", us.userID, err)
			}
		}
	}()

	// 1. Create user
	user, err := us.createUser()
	if err != nil {
		return fmt.Errorf("create user failed: %w", err)
	}
	userCreated = true // Mark user as created for cleanup

	// Small delay to simulate real user behavior
	time.Sleep(time.Duration(rand.Intn(300)+100) * time.Millisecond)

	// 2. Login
	if err := us.login(user); err != nil {
		journeyErr = fmt.Errorf("login failed: %w", err)
		return journeyErr
	}

	time.Sleep(time.Duration(rand.Intn(200)+100) * time.Millisecond)

	// 3. Create wordlist
	wordlistID, err := us.createWordlist()
	if err != nil {
		journeyErr = fmt.Errorf("create wordlist failed: %w", err)
		return journeyErr
	}

	time.Sleep(time.Duration(rand.Intn(200)+100) * time.Millisecond)

	// 4. Add words to wordlist
	if err := us.addWords(wordlistID); err != nil {
		journeyErr = fmt.Errorf("add words failed: %w", err)
		return journeyErr
	}

	time.Sleep(time.Duration(rand.Intn(300)+200) * time.Millisecond)

	// 5. Quiz loop - simulate multiple quiz sessions
	quizCount := rand.Intn(10) + 10 // 10-20 quizzes per user
	log.Printf("User %d: Starting quiz session (%d quizzes)", us.userID, quizCount)
	for i := 0; i < quizCount; i++ {
		if err := us.doQuiz(wordlistID); err != nil {
			// Log quiz error but continue with other quizzes
			log.Printf("User %d quiz %d failed: %v", us.userID, i+1, err)
		} else if i%5 == 0 { // Log every 5th quiz to avoid spam
			log.Printf("User %d: Completed quiz %d/%d", us.userID, i+1, quizCount)
		}

		// Random delay between quizzes
		time.Sleep(time.Duration(rand.Intn(500)+200) * time.Millisecond)
	}

	// Journey completed successfully
	return nil
}

func (us *UserSimulator) createUser() (User, error) {
	user := User{
		FirstName: gofakeit.FirstName(),
		LastName:  gofakeit.LastName(),
		Email:     fmt.Sprintf("loadtest_%d_%s@example.com", us.userID, gofakeit.UUID()),
		Password:  "LoadTest123!",
	}

	log.Printf("User %d: Creating user %s", us.userID, user.Email)
	body, _ := json.Marshal(user)
	req, _ := http.NewRequest("POST", us.config.BaseURL+"/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := us.client.Do(req, "POST /users")
	if err != nil {
		return user, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return user, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	log.Printf("User %d: Created user successfully (%d)", us.userID, resp.StatusCode)
	return user, nil
}

func (us *UserSimulator) login(user User) error {
	loginReq := LoginRequest{
		Email:    user.Email,
		Password: user.Password,
	}

	log.Printf("User %d: Logging in", us.userID)
	body, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest("POST", us.config.BaseURL+"/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := us.client.Do(req, "POST /login")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status: %d", resp.StatusCode)
	}

	// Extract auth token from headers
	us.authToken = resp.Header.Get("Authorization")
	if us.authToken == "" {
		return fmt.Errorf("no authorization token received")
	}

	log.Printf("User %d: Logged in successfully", us.userID)
	return nil
}

func (us *UserSimulator) createWordlist() (int, error) {
	wordlistReq := WordlistRequest{
		Name:         fmt.Sprintf("LoadTest Wordlist %d - %s", us.userID, gofakeit.Color()),
		LanguageCode: "en",
	}

	log.Printf("User %d: Creating wordlist '%s'", us.userID, wordlistReq.Name)
	body, _ := json.Marshal(wordlistReq)
	req, _ := http.NewRequest("POST", us.config.BaseURL+"/wordlists", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", us.authToken)

	resp, err := us.client.Do(req, "POST /wordlists")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("create wordlist failed with status: %d", resp.StatusCode)
	}

	var wordlistResp WordlistResponse
	if err := json.NewDecoder(resp.Body).Decode(&wordlistResp); err != nil {
		return 0, fmt.Errorf("failed to decode wordlist response: %w", err)
	}

	log.Printf("User %d: Created wordlist ID %d", us.userID, wordlistResp.ID)
	return wordlistResp.ID, nil
}

func (us *UserSimulator) addWords(wordlistID int) error {
	// Add 5-8 random words from our test set
	wordCount := rand.Intn(4) + 5
	selectedWords := make([]string, 0, wordCount)

	// Shuffle words and select a subset
	shuffled := make([]string, len(testWords))
	copy(shuffled, testWords)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	for i := 0; i < wordCount && i < len(shuffled); i++ {
		selectedWords = append(selectedWords, shuffled[i])
	}

	log.Printf("User %d: Adding %d words to wordlist %d", us.userID, len(selectedWords), wordlistID)
	for i, word := range selectedWords {
		wordReq := WordRequest{Name: word}
		body, _ := json.Marshal(wordReq)

		req, _ := http.NewRequest("POST",
			fmt.Sprintf("%s/wordlists/%d/words", us.config.BaseURL, wordlistID),
			bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", us.authToken)

		resp, err := us.client.Do(req, "POST /wordlists/{id}/words")
		if err != nil {
			return fmt.Errorf("failed to add word %s: %w", word, err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			return fmt.Errorf("add word %s failed with status: %d", word, resp.StatusCode)
		}

		if i%3 == 0 { // Log every 3rd word to avoid spam
			log.Printf("User %d: Added word '%s' (%d/%d)", us.userID, word, i+1, len(selectedWords))
		}

		// Small delay between word additions
		time.Sleep(time.Duration(rand.Intn(100)+50) * time.Millisecond)
	}

	return nil
}

func (us *UserSimulator) doQuiz(wordlistID int) error {
	// 1. Generate quiz
	req, _ := http.NewRequest("POST",
		fmt.Sprintf("%s/wordlists/%d/quizzes", us.config.BaseURL, wordlistID),
		bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", us.authToken)

	resp, err := us.client.Do(req, "POST /wordlists/{id}/quizzes")
	if err != nil {
		return fmt.Errorf("generate quiz failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("generate quiz failed with status: %d", resp.StatusCode)
	}

	var quizResp model.Quiz
	if err := json.NewDecoder(resp.Body).Decode(&quizResp); err != nil {
		return fmt.Errorf("failed to decode quiz response: %w", err)
	}

	// Simulate thinking time
	time.Sleep(time.Duration(rand.Intn(2000)+1000) * time.Millisecond)

	// 2. Save quiz result (simulate 80% success rate)
	success := rand.Float32() < 0.8
	responseTimeMs := rand.Intn(5000) + 1000 // 1-6 second response time
	
	saveReq := httpHandler.SaveInput{
		WordlistID:              int64(wordlistID),
		WordID:                  quizResp.WordID,
		DefinitionID:            quizResp.DefinitionID,
		LeitnerSystemTrackingID: quizResp.ID, // This is the LeitnerSystemID
		QuizType:                string(quizResp.Type),
		IsCorrect:               success,
		ResponseTimeMs:          responseTimeMs,
	}

	body, _ := json.Marshal(saveReq)
	req, _ = http.NewRequest("PATCH",
		fmt.Sprintf("%s/wordlists/%d/quizzes", us.config.BaseURL, wordlistID),
		bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", us.authToken)

	resp, err = us.client.Do(req, "PATCH /wordlists/{id}/quizzes")
	if err != nil {
		return fmt.Errorf("save quiz failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("save quiz failed with status: %d", resp.StatusCode)
	}

	return nil
}

func (us *UserSimulator) deleteUser() error {
	// Create request with context timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Printf("User %d: Cleaning up user account", us.userID)
	req, _ := http.NewRequestWithContext(ctx, "DELETE", us.config.BaseURL+"/users", nil)
	req.Header.Set("Authorization", us.authToken)

	resp, err := us.client.Do(req, "DELETE /users")
	if err != nil {
		return fmt.Errorf("delete request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete user failed with status: %d", resp.StatusCode)
	}

	log.Printf("User %d: Successfully cleaned up", us.userID)
	return nil
}

func printHelp() {
	fmt.Printf(`
🚀 Decorebator API Load Testing Tool

USAGE:
    go run tests/load_test_runner.go [FLAGS]
    make load-test ARGS="[FLAGS]"

FLAGS:
    -env string          Environment: 'local' or 'prod' (default: local)
    -url string          API base URL (overrides -env)
    -users int           Number of concurrent users (default: 1)
    -duration string     Test duration (e.g., 30s, 2m, 1h) (default: 30s)
    -timeout duration    HTTP client timeout (default: 30s)
    -help               Show this help message

ENVIRONMENTS:
    local               http://localhost:3000
    prod                https://api.decorebator.com

EXAMPLES:
    # Basic test with 1 user for 30 seconds
    make load-test ARGS="-env=local"

    # Test with 5 concurrent users for 1 minute
    make load-test ARGS="-users=5 -duration=1m -env=local"

    # Heavy load test with 20 users for 2 minutes
    make load-test ARGS="-users=20 -duration=2m -env=local"

    # Test against production
    make load-test ARGS="-users=10 -duration=30s -env=prod"

    # Custom URL
    make load-test ARGS="-url=http://localhost:8080 -users=5"

    # Quick smoke test
    make load-test ARGS="-users=1 -duration=5s"

WHAT IT TESTS:
    Complete user journey simulation:
    ✓ User Registration
    ✓ Authentication (JWT)
    ✓ Wordlist Creation
    ✓ Adding Words (5-10 random words)
    ✓ Quiz Loop (10-20 quizzes per user)
    ✓ Cleanup (automatic user deletion)

PERFORMANCE STANDARDS:
    🟢 Excellent: ≤ 100ms (Perceived as instantaneous)
    🟢 Good: ≤ 500ms (Good user experience)
    🟡 Acceptable: ≤ 1s (Acceptable for most apps)
    🟠 Poor: ≤ 2s (May impact user experience)
    🔴 Unacceptable: > 2s (Users likely to abandon)

`)
}

func main() {
	// Parse command line flags
	var config Config
	var durationStr string
	var environment string
	var showHelp bool

	flag.StringVar(&config.BaseURL, "url", "", "API base URL (overrides -env)")
	flag.StringVar(&environment, "env", "local", "Environment: 'local' or 'prod'")
	flag.IntVar(&config.Users, "users", 1, "Number of concurrent users")
	flag.StringVar(&durationStr, "duration", "30s", "Test duration (e.g., 30s, 2m)")
	flag.DurationVar(&config.ClientTimeout, "timeout", 30*time.Second, "HTTP client timeout")
	flag.BoolVar(&showHelp, "help", false, "Show help message")
	flag.Parse()

	// Show help if requested or if no arguments provided and not using defaults
	if showHelp || (len(os.Args) == 1) {
		printHelp()
		return
	}

	// Set base URL based on environment if not explicitly provided
	if config.BaseURL == "" {
		switch environment {
		case "local":
			config.BaseURL = "http://localhost:3000"
		case "prod":
			config.BaseURL = "https://api.decorebator.com"
		default:
			log.Fatalf("Invalid environment '%s'. Use 'local' or 'prod'", environment)
		}
	}

	var err error
	config.Duration, err = time.ParseDuration(durationStr)
	if err != nil {
		log.Fatalf("Invalid duration format: %v", err)
	}

	fmt.Printf("=== Decorebator Load Test ===\n")
	fmt.Printf("Base URL: %s\n", config.BaseURL)
	fmt.Printf("Concurrent Users: %d\n", config.Users)
	fmt.Printf("Duration: %s\n", config.Duration)
	fmt.Printf("Client Timeout: %s\n", config.ClientTimeout)
	fmt.Printf("\nStarting load test...\n\n")

	// Initialize metrics collector
	collector := &MetricsCollector{}

	// Create HTTP client with metrics
	httpClient := &http.Client{
		Timeout: config.ClientTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 30,
			IdleConnTimeout:     90 * time.Second,
		},
	}

	metricsClient := &MetricsHTTPClient{
		client:    httpClient,
		collector: collector,
	}

	// Setup goroutines and wait group
	var wg sync.WaitGroup
	startTime := time.Now()

	// Channel to signal completion
	done := make(chan struct{})

	// Start timer to stop test after duration
	go func() {
		time.Sleep(config.Duration)
		close(done)
	}()

	// Launch user simulators
	for i := 0; i < config.Users; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			simulator := NewUserSimulator(metricsClient, config, userID)

			// Keep running user journeys until done signal
			journeyCount := 0
			for {
				select {
				case <-done:
					return
				default:
					if err := simulator.SimulateUserJourney(); err != nil {
						log.Printf("User %d journey %d failed: %v", userID, journeyCount+1, err)
					}
					journeyCount++

					// Small delay before starting next journey
					time.Sleep(time.Duration(rand.Intn(1000)+500) * time.Millisecond)
				}
			}
		}(i)
	}

	// Wait for all users to complete
	wg.Wait()

	// Calculate test duration
	testDuration := time.Since(startTime)

	// Generate and display results
	fmt.Printf("=== Load Test Results ===\n")
	fmt.Printf("Test Duration: %s\n", testDuration.Round(time.Millisecond))
	fmt.Printf("Concurrent Users: %d\n", config.Users)
	fmt.Printf("\n")

	stats := collector.GetStats()

	totalRequests := 0
	totalSuccesses := 0
	totalErrors := 0

	// Display endpoint statistics with performance ratings
	fmt.Printf("Endpoint Performance:\n")
	overallWorst := time.Duration(0)
	for _, stat := range stats {
		successRate := float64(stat.Success) / float64(stat.Total) * 100
		
		avgRating, _ := getPerformanceRating(stat.AvgDuration)
		p95Rating, _ := getPerformanceRating(stat.P95Duration)
		
		fmt.Printf("- %-30s %4d requests, %2d errors, avg: %6s %s, p95: %6s %s, success: %5.1f%%\n",
			stat.Endpoint+":",
			stat.Total,
			stat.Errors,
			stat.AvgDuration.Round(time.Millisecond),
			avgRating,
			stat.P95Duration.Round(time.Millisecond),
			p95Rating,
			successRate)

		// Track worst performance for overall assessment
		if stat.P95Duration > overallWorst {
			overallWorst = stat.P95Duration
		}

		totalRequests += stat.Total
		totalSuccesses += stat.Success
		totalErrors += stat.Errors
	}

	// Overall performance assessment
	overallRating, overallDesc := getPerformanceRating(overallWorst)
	fmt.Printf("\n📊 Overall Performance Assessment: %s\n", overallRating)
	fmt.Printf("   %s (based on worst P95: %s)\n", overallDesc, overallWorst.Round(time.Millisecond))

	// Overall statistics
	fmt.Printf("\nOverall Statistics:\n")
	fmt.Printf("- Total Requests: %d\n", totalRequests)
	fmt.Printf("- Success Rate: %.1f%%\n", float64(totalSuccesses)/float64(totalRequests)*100)
	fmt.Printf("- Error Rate: %.1f%%\n", float64(totalErrors)/float64(totalRequests)*100)
	fmt.Printf("- Average RPS: %.1f\n", float64(totalRequests)/testDuration.Seconds())

	// Error breakdown
	if totalErrors > 0 {
		fmt.Printf("\nError Breakdown:\n")
		for endpoint, stat := range stats {
			if stat.Errors > 0 {
				fmt.Printf("- %s:\n", endpoint)
				for errorMsg, count := range stat.ErrorMessages {
					fmt.Printf("  * %s: %d times\n", errorMsg, count)
				}
			}
		}
	}

	fmt.Printf("\n📋 Performance Standards Reference:\n")
	fmt.Printf("   🟢 Excellent: ≤ %s (Perceived as instantaneous)\n", thresholds.Excellent)
	fmt.Printf("   🟢 Good: ≤ %s (Good user experience)\n", thresholds.Good)
	fmt.Printf("   🟡 Acceptable: ≤ %s (Acceptable for most apps)\n", thresholds.Acceptable)
	fmt.Printf("   🟠 Poor: ≤ %s (May impact user experience)\n", thresholds.Poor)
	fmt.Printf("   🔴 Unacceptable: > %s (Users likely to abandon)\n", thresholds.Poor)

	fmt.Printf("\nLoad test completed successfully!\n")
}
