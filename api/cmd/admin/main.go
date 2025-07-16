package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultLocalURL = "http://localhost:3000"
	defaultProdURL  = "https://api.decorebator.com" // Update this with your actual production URL
)

var (
	env     = flag.String("env", "local", "Environment: local or prod")
	token   = flag.String("token", "", "Static authentication token (overrides env vars)")
	baseURL = flag.String("url", "", "Base URL (overrides default URLs)")
	timeout = flag.Int("timeout", 30, "Request timeout in seconds")
	output  = flag.String("output", "", "Output file for profile downloads")
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] <command>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Admin tool for triggering pprof and monitoring endpoints.\n\n")
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  health              Check system health\n")
		fmt.Fprintf(os.Stderr, "  metrics             Get system metrics\n")
		fmt.Fprintf(os.Stderr, "  info                Get system information\n")
		fmt.Fprintf(os.Stderr, "  pprof-index         Get pprof index page\n")
		fmt.Fprintf(os.Stderr, "  pprof-cpu           Download CPU profile\n")
		fmt.Fprintf(os.Stderr, "  pprof-heap          Download heap profile\n")
		fmt.Fprintf(os.Stderr, "  pprof-goroutine     Download goroutine profile\n")
		fmt.Fprintf(os.Stderr, "  pprof-block         Download block profile\n")
		fmt.Fprintf(os.Stderr, "  pprof-mutex         Download mutex profile\n")
		fmt.Fprintf(os.Stderr, "  pprof-trace         Download execution trace\n")
		fmt.Fprintf(os.Stderr, "  status              Check all endpoints\n")
		fmt.Fprintf(os.Stderr, "\nFlags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -env=local health\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -env=prod -token=your-token pprof-cpu\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -url=http://staging.example.com -token=token metrics\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -env=local -output=./profiles/cpu.prof pprof-cpu\n", os.Args[0])
	}

	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	command := flag.Arg(0)

	// Determine base URL
	url := getBaseURL()

	// Get authentication token
	authToken := getAuthToken()
	if authToken == "" {
		fmt.Fprintf(os.Stderr, "Error: No authentication token provided.\n")
		fmt.Fprintf(os.Stderr, "Use -token flag or set STATIC_AUTHENTICATION environment variable.\n")
		os.Exit(1)
	}

	client := &http.Client{
		Timeout: time.Duration(*timeout) * time.Second,
	}

	switch command {
	case "health":
		makeRequest(client, url+"/static/admin/health", authToken, false)
	case "metrics":
		makeRequest(client, url+"/static/admin/metrics", authToken, false)
	case "info":
		makeRequest(client, url+"/static/admin/info", authToken, false)
	case "pprof-index":
		makeRequest(client, url+"/static/admin/debug/pprof/", authToken, false)
	case "pprof-cpu":
		filename := getOutputFilename("cpu-profile", "prof")
		downloadProfile(client, url+"/static/admin/debug/pprof/profile", authToken, filename)
	case "pprof-heap":
		filename := getOutputFilename("heap-profile", "prof")
		downloadProfile(client, url+"/static/admin/debug/pprof/heap", authToken, filename)
	case "pprof-goroutine":
		filename := getOutputFilename("goroutine-profile", "prof")
		downloadProfile(client, url+"/static/admin/debug/pprof/goroutine", authToken, filename)
	case "pprof-block":
		filename := getOutputFilename("block-profile", "prof")
		downloadProfile(client, url+"/static/admin/debug/pprof/block", authToken, filename)
	case "pprof-mutex":
		filename := getOutputFilename("mutex-profile", "prof")
		downloadProfile(client, url+"/static/admin/debug/pprof/mutex", authToken, filename)
	case "pprof-trace":
		filename := getOutputFilename("trace", "out")
		downloadProfile(client, url+"/static/admin/debug/pprof/trace", authToken, filename)
	case "status":
		fmt.Println("=== Admin Endpoints Status ===")
		fmt.Printf("Environment: %s\n", *env)
		fmt.Printf("Base URL: %s\n", url)
		fmt.Println()

		fmt.Println("--- Health Check ---")
		makeRequest(client, url+"/static/admin/health", authToken, false)
		fmt.Println()

		fmt.Println("--- System Info ---")
		makeRequest(client, url+"/static/admin/info", authToken, false)
		fmt.Println()

		fmt.Println("--- Metrics ---")
		makeRequest(client, url+"/static/admin/metrics", authToken, false)
		fmt.Println()

		fmt.Println("--- pprof Index ---")
		makeRequest(client, url+"/static/admin/debug/pprof/", authToken, false)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		flag.Usage()
		os.Exit(1)
	}
}

func getBaseURL() string {
	if *baseURL != "" {
		return strings.TrimSuffix(*baseURL, "/")
	}

	switch *env {
	case "local":
		return defaultLocalURL
	case "prod":
		return defaultProdURL
	default:
		fmt.Fprintf(os.Stderr, "Invalid environment: %s (must be 'local' or 'prod')\n", *env)
		os.Exit(1)
		return ""
	}
}

func getAuthToken() string {
	if *token != "" {
		return *token
	}

	// Try to get from environment variable
	return os.Getenv("STATIC_AUTHENTICATION")
}

func getOutputFilename(prefix, extension string) string {
	if *output != "" {
		return *output
	}

	timestamp := time.Now().Format("20060102-150405")
	return fmt.Sprintf("%s-%s.%s", prefix, timestamp, extension)
}

func makeRequest(client *http.Client, url, authToken string, isDownload bool) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
		return
	}

	req.Header.Set("Authorization", authToken)
	if !isDownload {
		req.Header.Set("Content-Type", "application/json")
	}

	fmt.Printf("Making request to: %s\n", url)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error making request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d %s\n", resp.StatusCode, resp.Status)

	if resp.StatusCode >= 400 {
		fmt.Fprintf(os.Stderr, "Request failed with status %d\n", resp.StatusCode)
	}

	// Print response body for non-download requests
	if !isDownload {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
			return
		}
		fmt.Println(string(body))
	}
}

func downloadProfile(client *http.Client, url, authToken, filename string) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
		return
	}

	req.Header.Set("Authorization", authToken)

	fmt.Printf("Downloading profile from: %s\n", url)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error making request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Download failed with status %d\n", resp.StatusCode)
		return
	}

	// Create output directory if it doesn't exist
	if dir := filepath.Dir(filename); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
			return
		}
	}

	// Create output file
	file, err := os.Create(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
		return
	}
	defer file.Close()

	// Copy response body to file
	size, err := io.Copy(file, resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing to file: %v\n", err)
		return
	}

	fmt.Printf("Profile saved to: %s (%d bytes)\n", filename, size)
	fmt.Printf("Analyze with: go tool pprof %s\n", filename)
}
