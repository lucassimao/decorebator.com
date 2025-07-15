# PostgreSQL Connection Benchmarking Tool

A realistic database connection benchmarking tool that tests PostgreSQL connection limits using the actual complex quiz generation query from the Decorebator application.

## Features

- **Realistic Load Testing**: Uses the actual complex quiz generation query with multiple CTEs, joins, and aggregations
- **Progressive Testing**: Tests connection levels from 1 to 1000 concurrent connections
- **No Connection Pooling**: Uses raw `pgx.Connect()` calls to test database limits
- **Comprehensive Metrics**: Tracks success rates, response times, and error categorization
- **Memory Monitoring**: Shows memory usage at each concurrency level
- **Production Query**: Benchmarks with the actual Leitner spaced repetition algorithm query

## Usage

### Basic Usage
```bash
cd api
go run cmd/benchmark/db_connections.go -url="postgresql://user:pass@host:port/database"
```

### With Custom Parameters
```bash
# Test up to 500 concurrent connections with 15-second timeout
go run cmd/benchmark/db_connections.go \
  -url="postgresql://user:pass@host:port/database" \
  -max=500 \
  -timeout=15
```

### Connection Pool Testing
```bash
# Test with pgxpool.Pool (recommended for production comparison)
go run cmd/benchmark/db_connections.go \
  -url="postgresql://user:pass@host:port/database" \
  -pool \
  -max-pool-size=20 \
  -min-pool-size=2 \
  -max=100

# Compare pooled vs non-pooled performance
go run cmd/benchmark/db_connections.go -url="your_url" -max=50                    # No pooling
go run cmd/benchmark/db_connections.go -url="your_url" -max=50 -pool             # With pooling
```

### Parameters

- `-url`: PostgreSQL connection URL (required)
- `-max`: Maximum concurrent connections to test (default: 1000)
- `-timeout`: Connection timeout in seconds (default: 10)
- `-pool`: Use pgxpool.Pool instead of individual connections (default: false)
- `-max-pool-size`: Maximum pool size when using -pool (default: 20)
- `-min-pool-size`: Minimum pool size when using -pool (default: 2)

## Query Details

The benchmark uses the actual quiz generation query from `leitner_system_strategy.go`, which includes:

- **Multiple CTEs**: Complex WITH clauses for definition priorities and weighting
- **Multiple JOINs**: Joins across 6+ tables (leitner_system_tracking, definitions, words, etc.)
- **JSON Aggregation**: Aggregates example audio files into JSON arrays
- **Complex Calculations**: Leitner box intervals, progress ratios, and priority weights
- **Realistic Parameters**: Uses sample userID=1 and wordlistID=1

This provides a much more realistic benchmark than simple `SELECT 1` queries.

## Connection Modes

### Individual Connections (Default)
- Each test goroutine creates its own `pgx.Connect()` connection
- Tests raw database connection limits
- Simulates worst-case scenario where no connection pooling is used
- Each connection goes through full TCP handshake and authentication

### Connection Pool Mode (`-pool`)
- Uses `pgxpool.Pool` with configurable min/max connections
- Connections are reused across requests
- Simulates production environment with proper connection pooling
- More efficient resource usage and typically better performance
- Matches the actual production configuration from `database.go`

### Pool Configuration
When using `-pool`, the benchmark uses production-like settings:
- `MaxConnLifetime`: 1 hour
- `MaxConnIdleTime`: 15 minutes
- `ConnectTimeout`: Uses the `-timeout` parameter
- Pool size controlled by `-max-pool-size` and `-min-pool-size`

## Sample Output

```
PostgreSQL Connection Benchmarking Tool - Decorebator Quiz Generation
====================================================================
Database URL: postgresql://user:****@host:port/database
Max Concurrency: 1000
Connection Timeout: 10 seconds
Query Type: Complex Quiz Generation (production-realistic)

Testing single connection...
✓ Single connection successful

Starting concurrent connection tests...

Testing 1 concurrent connections...
✓ Success: 1/1 connections (avg: 45ms, min: 45ms, max: 45ms)
Memory usage: 12.34 MB -> 12.56 MB (diff: 0.22 MB)

Testing 5 concurrent connections...
✓ Success: 5/5 connections (avg: 52ms, min: 41ms, max: 67ms)
Memory usage: 12.56 MB -> 13.21 MB (diff: 0.65 MB)

Testing 100 concurrent connections...
⚠️  Partial: 85/100 connections (avg: 1.2s, min: 234ms, max: 3.4s)
Errors: too_many_connections (15)
Memory usage: 25.67 MB -> 45.23 MB (diff: 19.56 MB)

Summary
=======
Maximum successful concurrent connections: 50
First failure at: 100 concurrent connections (85 succeeded, 15 failed)

Performance Summary:
Concurrency | Success | Avg Duration | Total Duration
------------|---------|--------------|---------------
          1 |       1 |         45ms |        125ms
          5 |       5 |         52ms |        287ms
         10 |      10 |         78ms |        445ms
         50 |      50 |        234ms |       1.2s
        100 |      85 |        1.2s |       8.7s
```

## Error Categories

The tool categorizes database errors:

- `connection_refused`: Database server refused connection
- `timeout`: Connection or query timeout
- `too_many_connections`: Database max_connections limit reached
- `auth_failed`: Authentication failed
- `database_not_found`: Database does not exist
- `host_not_found`: Host not reachable
- `other`: Other database errors

## Next Steps

This tool provides the baseline for connection benchmarking. The next version will include:

1. **pgxpool.Pool Testing**: Compare pooled vs non-pooled performance
2. **Pool Configuration Testing**: Test different MaxConns values
3. **Connection Reuse Metrics**: Measure pool efficiency
4. **Throughput Testing**: Queries per second measurements

## Requirements

- Go 1.23+
- PostgreSQL 15+ database with Decorebator schema
- Database must contain the required tables: `leitner_system_tracking`, `definitions`, `words`, `wordlists`