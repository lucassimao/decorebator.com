# Decorebator API Load Testing

Simple load testing script that simulates realistic user behavior to test your API infrastructure under concurrent load.

## What It Tests

The script simulates a complete user journey:

1. **User Registration** - Creates a new user account
2. **Authentication** - Logs in and gets JWT token  
3. **Wordlist Creation** - Creates a new wordlist
4. **Add Words** - Adds 5-10 predefined words to the wordlist
5. **Quiz Loop** - Generates and answers 10-20 quizzes per user
6. **Cleanup** - Deletes the user and all associated data (wordlists, words, definitions, progress, etc.)

## Usage

### Basic Usage

```bash
# Run with default settings (1 user, 30 seconds, local environment)
cd api
go run tests/load_test_runner.go -words=tests/words.txt

# Test against local development server (explicit)
go run tests/load_test_runner.go -env=local -words=tests/words.txt

# Test against production API  
go run tests/load_test_runner.go -env=prod -words=tests/words.txt

# Override with custom URL
go run tests/load_test_runner.go -url=http://localhost:8080 -words=tests/words.txt
```

### Configuration Options

```bash
# Single user test against local server (default)
go run tests/load_test_runner.go

# 5 concurrent users for 1 minute against local
go run tests/load_test_runner.go -users=5 -duration=1m -env=local

# 20 concurrent users for 2 minutes against production
go run tests/load_test_runner.go -users=20 -duration=2m -env=prod

# Heavy load test: 50 users for 5 minutes against production
go run tests/load_test_runner.go -users=50 -duration=5m -env=prod

# Quick smoke test: single user for 10 seconds
go run tests/load_test_runner.go -users=1 -duration=10s
```

### Available Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-env` | `local` | Environment: 'local' or 'prod' |
| `-url` | (auto-set) | API base URL (overrides -env) |
| `-users` | `1` | Number of concurrent users |
| `-duration` | `30s` | Test duration (e.g., 30s, 2m, 1h) |
| `-timeout` | `30s` | HTTP client timeout |

**Environment URLs:**
- `local`: `http://localhost:3000`
- `prod`: `https://api.decorebator.com`

## Sample Output

```
=== Decorebator Load Test ===
Base URL: http://localhost:8080
Concurrent Users: 10
Duration: 30s
Client Timeout: 30s

Starting load test...

=== Load Test Results ===
Test Duration: 30.2s
Concurrent Users: 10

Endpoint Performance:
- POST /users:                  10 requests,  0 errors, avg:  245ms, p95:  412ms, success:  100.0%
- POST /login:                  10 requests,  0 errors, avg:  156ms, p95:  289ms, success:  100.0%
- POST /wordlists:              10 requests,  0 errors, avg:  189ms, p95:  334ms, success:  100.0%
- POST /wordlists/{id}/words:   67 requests,  1 errors, avg:  167ms, p95:  456ms, success:   98.5%
- POST /wordlists/{id}/quizzes: 142 requests, 2 errors, avg:  334ms, p95:  678ms, success:   98.6%
- PATCH /wordlists/{id}/quizzes: 140 requests, 0 errors, avg:  123ms, p95:  267ms, success:  100.0%
- DELETE /users:                10 requests,  0 errors, avg:   89ms, p95:  156ms, success:  100.0%

Overall Statistics:
- Total Requests: 379
- Success Rate: 99.2%
- Error Rate: 0.8%
- Average RPS: 12.5

Load test completed successfully!
```

## Understanding the Results

### Key Metrics

- **avg**: Average response time for the endpoint
- **p95**: 95th percentile response time (95% of requests were faster than this)
- **success**: Percentage of successful requests
- **RPS**: Requests per second

### Performance Targets

Good performance indicators:
- **Success Rate**: > 99%
- **Average Response Time**: < 500ms for most endpoints
- **P95 Response Time**: < 1000ms for most endpoints
- **Quiz Generation**: < 800ms avg (this is CPU intensive)

### Warning Signs

Watch out for:
- Success rate < 95%
- P95 response times > 2000ms
- High error rates on specific endpoints
- Declining RPS over time (memory leaks)

## Test Data

The script uses:
- **Unique emails** for each user (format: `loadtest_{id}_{uuid}@example.com`)
- **Random names** generated using gofakeit
- **Predefined word list**: hello, world, computer, language, study, learn, practice, memory, quiz, vocabulary, excellent, beautiful, important, different, necessary
- **Realistic delays** between requests (100-500ms)
- **80% quiz success rate** to simulate real user behavior
- **Guaranteed cleanup** using defer statements - ensures all test users are deleted even if the journey fails partway through, preventing database bloating

## Before Running Tests

1. **Start your API server**:
   ```bash
   cd api
   make watch  # or make run
   ```

2. **Ensure dependencies are running**:
   ```bash
   docker-compose up -d  # PostgreSQL, Redis, MinIO
   ```

3. **Apply database migrations**:
   ```bash
   make migrate-up
   ```

## Troubleshooting

### Common Issues

**Connection refused errors**:
- Check if the API server is running
- Verify the correct URL with `-url` flag

**High error rates**:
- Database connection limits might be reached
- Check server logs for specific errors
- Reduce concurrent users or test duration

**Timeout errors**:
- Increase timeout with `-timeout=60s`
- Check if server is overloaded

**Database errors**:
- Ensure PostgreSQL is running and accessible
- Check connection pool limits in your API configuration

### Performance Tips

**For reliable results**:
- Run tests multiple times and average results
- Start with low concurrent users (2-5) and increase gradually
- Monitor server resources (CPU, memory, database connections)
- Test against a staging environment similar to production

**Scaling recommendations**:
- If handling 10 concurrent users easily, try 20, then 50
- Target should be comfortable handling 2-3x your expected peak load
- Consider both sustained load and burst traffic patterns

## Integration with Development Workflow

```bash
# Quick health check
go run tests/load_test_runner.go -users=2 -duration=10s

# Before deployment
go run tests/load_test_runner.go -users=20 -duration=2m

# Capacity planning
go run tests/load_test_runner.go -users=50 -duration=5m
```

## Dependencies

Uses existing dependencies from `go.mod`:
- `github.com/brianvoe/gofakeit/v7` - Fake data generation
- Standard library packages for HTTP, concurrency, and JSON

No external load testing tools required!