package http

import (
	"net/http"
	"net/http/pprof"
	"os"
	"runtime"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"

	"decorebator.com/internal/app"
	"decorebator.com/internal/common"
	"github.com/gin-gonic/gin"
)

// PprofHandler wraps the standard pprof endpoints with logging
func PprofHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log admin access for audit trail
		common.Logger.Info("admin pprof access",
			"endpoint", c.Request.URL.Path,
			"method", c.Request.Method,
			"remote_addr", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)

		// Extract the profile type from the URL
		profileType := c.Param("profile")
		// Remove leading slash if present (wildcard routes include the slash)
		profileType = strings.TrimPrefix(profileType, "/")

		// Handle different pprof endpoints
		switch profileType {
		case "profile":
			// CPU profiling with configurable duration (default 30s)
			// Set content type for binary profile
			c.Header("Content-Type", "application/octet-stream")
			c.Header("Content-Disposition", "attachment; filename=profile.prof")

			// Use standard pprof.Profile handler
			pprof.Profile(c.Writer, c.Request)
		case "heap":
			c.Header("Content-Type", "application/octet-stream")
			c.Header("Content-Disposition", "attachment; filename=heap.prof")
			pprof.Handler("heap").ServeHTTP(c.Writer, c.Request)
		case "goroutine":
			c.Header("Content-Type", "application/octet-stream")
			c.Header("Content-Disposition", "attachment; filename=goroutine.prof")
			pprof.Handler("goroutine").ServeHTTP(c.Writer, c.Request)
		case "block":
			c.Header("Content-Type", "application/octet-stream")
			c.Header("Content-Disposition", "attachment; filename=block.prof")
			pprof.Handler("block").ServeHTTP(c.Writer, c.Request)
		case "mutex":
			c.Header("Content-Type", "application/octet-stream")
			c.Header("Content-Disposition", "attachment; filename=mutex.prof")
			pprof.Handler("mutex").ServeHTTP(c.Writer, c.Request)
		case "allocs":
			c.Header("Content-Type", "application/octet-stream")
			c.Header("Content-Disposition", "attachment; filename=allocs.prof")
			pprof.Handler("allocs").ServeHTTP(c.Writer, c.Request)
		case "trace":
			// Execution trace with configurable duration (default 1s)
			c.Header("Content-Type", "application/octet-stream")
			c.Header("Content-Disposition", "attachment; filename=trace.out")
			pprof.Trace(c.Writer, c.Request)
		case "cmdline":
			c.Header("Content-Type", "text/plain")
			pprof.Cmdline(c.Writer, c.Request)

		case "symbol":
			c.Header("Content-Type", "text/plain")
			pprof.Symbol(c.Writer, c.Request)

		case "", "/":
			// Index page
			c.Header("Content-Type", "text/html")
			pprof.Index(c.Writer, c.Request)
		default:
			c.JSON(http.StatusNotFound, gin.H{"error": "unknown profile type"})
		}
	}
}

// HealthCheckHandler provides comprehensive system health information
func HealthCheckHandler(appCtx *app.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log health check access
		common.Logger.Info("admin health check access",
			"remote_addr", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)

		health := gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"services":  gin.H{},
			"system":    getSystemInfo(),
		}

		// Check database health
		if appCtx.Database != nil {
			start := time.Now()
			err := appCtx.Database.Ping(c.Request.Context())
			latency := time.Since(start)

			services, ok := health["services"].(gin.H)
			if !ok {
				services = gin.H{}
				health["services"] = services
			}
			if err != nil {
				services["database"] = gin.H{
					"status": "unhealthy",
					"error":  err.Error(),
				}
				health["status"] = "unhealthy"
			} else {
				services["database"] = gin.H{
					"status":  "healthy",
					"latency": latency.String(),
				}
			}
		}

		// Check Redis health
		if appCtx.RedisClient != nil {
			start := time.Now()
			_, err := appCtx.RedisClient.Ping(c.Request.Context()).Result()
			latency := time.Since(start)

			services, ok := health["services"].(gin.H)
			if !ok {
				services = gin.H{}
				health["services"] = services
			}
			if err != nil {
				services["redis"] = gin.H{
					"status": "unhealthy",
					"error":  err.Error(),
				}
				health["status"] = "unhealthy"
			} else {
				services["redis"] = gin.H{
					"status":  "healthy",
					"latency": latency.String(),
				}
			}
		}

		// Return appropriate status code
		if health["status"] == "healthy" {
			c.JSON(http.StatusOK, health)
		} else {
			c.JSON(http.StatusServiceUnavailable, health)
		}
	}
}

// MetricsHandler provides performance metrics
func MetricsHandler(appCtx *app.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log metrics access
		common.Logger.Info("admin metrics access",
			"remote_addr", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)

		metrics := gin.H{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"system":    getSystemInfo(),
			"runtime":   getRuntimeMetrics(),
		}

		// Get database connection stats if available
		if appCtx.Database != nil {
			stats := appCtx.Database.Stat()
			metrics["database"] = gin.H{
				"total_conns":                stats.TotalConns(),
				"acquired_conns":             stats.AcquiredConns(),
				"idle_conns":                 stats.IdleConns(),
				"constructing_conns":         stats.ConstructingConns(),
				"max_conns":                  stats.MaxConns(),
				"acquire_count":              stats.AcquireCount(),
				"acquire_duration":           stats.AcquireDuration().String(),
				"canceled_acquire_count":     stats.CanceledAcquireCount(),
				"empty_acquire_count":        stats.EmptyAcquireCount(),
				"empty_acquire_wait_time":    stats.EmptyAcquireWaitTime().String(),
				"new_conns_count":            stats.NewConnsCount(),
				"max_idle_destroy_count":     stats.MaxIdleDestroyCount(),
				"max_lifetime_destroy_count": stats.MaxLifetimeDestroyCount(),
			}
		}

		c.JSON(http.StatusOK, metrics)
	}
}

// SystemInfoHandler provides build and version information
func SystemInfoHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log system info access
		common.Logger.Info("admin system info access",
			"remote_addr", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)

		info := gin.H{
			"timestamp":   time.Now().UTC().Format(time.RFC3339),
			"go_version":  runtime.Version(),
			"go_os":       runtime.GOOS,
			"go_arch":     runtime.GOARCH,
			"environment": os.Getenv("ENV"),
			"system":      getSystemInfo(),
		}

		c.JSON(http.StatusOK, info)
	}
}

// getSystemInfo returns current system statistics
func getSystemInfo() gin.H {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return gin.H{
		"memory": gin.H{
			"alloc":         formatBytes(m.Alloc),
			"total_alloc":   formatBytes(m.TotalAlloc),
			"sys":           formatBytes(m.Sys),
			"heap_alloc":    formatBytes(m.HeapAlloc),
			"heap_sys":      formatBytes(m.HeapSys),
			"heap_idle":     formatBytes(m.HeapIdle),
			"heap_inuse":    formatBytes(m.HeapInuse),
			"heap_released": formatBytes(m.HeapReleased),
			"heap_objects":  m.HeapObjects,
			"stack_inuse":   formatBytes(m.StackInuse),
			"stack_sys":     formatBytes(m.StackSys),
			"gc_cycles":     m.NumGC,
			"gc_pause":      formatGCPause(m.PauseNs, m.NumGC),
		},
		"goroutines": runtime.NumGoroutine(),
		"cpu_cores":  runtime.NumCPU(),
	}
}

// getRuntimeMetrics returns comprehensive runtime performance metrics
func getRuntimeMetrics() gin.H {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Get GC stats
	var gcStats debug.GCStats
	debug.ReadGCStats(&gcStats)

	// Calculate GC pause percentiles
	pausePercentiles := calculateGCPausePercentiles(gcStats.Pause)

	// Get goroutine stack analysis
	goroutineBreakdown := analyzeGoroutineStates()

	return gin.H{
		"memory_detailed": gin.H{
			"heap_in_use":     formatBytes(memStats.HeapInuse),
			"heap_idle":       formatBytes(memStats.HeapIdle),
			"heap_released":   formatBytes(memStats.HeapReleased),
			"heap_sys":        formatBytes(memStats.HeapSys),
			"next_gc":         formatBytes(memStats.NextGC),
			"last_gc":         formatLastGCTime(memStats.LastGC),
			"gc_cpu_fraction": memStats.GCCPUFraction,
			"num_forced_gc":   memStats.NumForcedGC,
			"mallocs":         memStats.Mallocs,
			"frees":           memStats.Frees,
			"heap_objects":    memStats.HeapObjects,
		},
		"gc_detailed": gin.H{
			"num_gc":          gcStats.NumGC,
			"pause_total":     gcStats.PauseTotal.String(),
			"pause_50th":      pausePercentiles["50th"],
			"pause_95th":      pausePercentiles["95th"],
			"pause_99th":      pausePercentiles["99th"],
			"pause_max":       pausePercentiles["max"],
			"gc_cpu_fraction": memStats.GCCPUFraction,
			"last_gc_time":    formatLastGCTime(memStats.LastGC),
		},
		"goroutine_analysis": goroutineBreakdown,
		"runtime_info": gin.H{
			"go_version":   runtime.Version(),
			"go_maxprocs":  runtime.GOMAXPROCS(0),
			"num_cgo_call": runtime.NumCgoCall(),
		},
	}
}

// calculateGCPausePercentiles calculates GC pause percentiles
func calculateGCPausePercentiles(pauses []time.Duration) gin.H {
	if len(pauses) == 0 {
		return gin.H{
			"50th": "0s",
			"95th": "0s",
			"99th": "0s",
			"max":  "0s",
		}
	}

	// Filter out zero values and sort
	var validPauses []time.Duration
	for _, pause := range pauses {
		if pause > 0 {
			validPauses = append(validPauses, pause)
		}
	}

	if len(validPauses) == 0 {
		return gin.H{
			"50th": "0s",
			"95th": "0s",
			"99th": "0s",
			"max":  "0s",
		}
	}

	sort.Slice(validPauses, func(i, j int) bool {
		return validPauses[i] < validPauses[j]
	})

	n := len(validPauses)
	return gin.H{
		"50th": validPauses[n*50/100].String(),
		"95th": validPauses[n*95/100].String(),
		"99th": validPauses[n*99/100].String(),
		"max":  validPauses[n-1].String(),
	}
}

// analyzeGoroutineStates analyzes goroutine states for debugging
func analyzeGoroutineStates() gin.H {
	// Get stack trace of all goroutines
	buf := make([]byte, 1024*1024) // 1MB buffer
	n := runtime.Stack(buf, true)
	stackTrace := string(buf[:n])

	// Count goroutines by state
	blocked := strings.Count(stackTrace, "goroutine") - strings.Count(stackTrace, "[running]")
	running := strings.Count(stackTrace, "[running]")

	// More detailed analysis
	networkBlocked := strings.Count(stackTrace, "net.")
	chanBlocked := strings.Count(stackTrace, "chan receive") + strings.Count(stackTrace, "chan send")
	ioBlocked := strings.Count(stackTrace, "internal/poll.") + strings.Count(stackTrace, "syscall.")
	sqlBlocked := strings.Count(stackTrace, "database/sql") + strings.Count(stackTrace, "github.com/jackc/pgx")

	return gin.H{
		"total_goroutines": runtime.NumGoroutine(),
		"running":          running,
		"blocked":          blocked,
		"network_blocked":  networkBlocked,
		"channel_blocked":  chanBlocked,
		"io_blocked":       ioBlocked,
		"sql_blocked":      sqlBlocked,
		"stack_summary":    summarizeTopStacks(stackTrace),
	}
}

// summarizeTopStacks summarizes the most common goroutine stack traces
func summarizeTopStacks(stackTrace string) []gin.H {
	// Split by goroutine entries
	goroutines := strings.Split(stackTrace, "goroutine ")
	stackCounts := make(map[string]int)

	for _, goroutine := range goroutines {
		if len(goroutine) < 10 {
			continue
		}

		// Extract the first meaningful line of the stack
		lines := strings.Split(goroutine, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "(") && !strings.Contains(line, "goroutine ") {
				// This is a function call line
				if strings.Contains(line, "decorebator.com/") {
					// Only count our application code
					stackCounts[line]++
				}
				break
			}
		}
	}

	// Convert to sorted slice
	type stackCount struct {
		Function string `json:"function"`
		Count    int    `json:"count"`
	}

	var stacks []stackCount
	for function, count := range stackCounts {
		stacks = append(stacks, stackCount{Function: function, Count: count})
	}

	// Sort by count, descending
	sort.Slice(stacks, func(i, j int) bool {
		return stacks[i].Count > stacks[j].Count
	})

	// Return top 5
	var result []gin.H
	for i, stack := range stacks {
		if i >= 5 {
			break
		}
		result = append(result, gin.H{
			"function": stack.Function,
			"count":    stack.Count,
		})
	}

	return result
}

// formatBytes converts bytes to human readable format
func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return strconv.FormatUint(b, 10) + " B"
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return strconv.FormatFloat(float64(b)/float64(div), 'f', 1, 64) + " " + "KMGTPE"[exp:exp+1] + "B"
}

// formatGCPause safely formats GC pause time to avoid integer overflow
func formatGCPause(pauseNs [256]uint64, numGC uint32) string {
	if numGC == 0 {
		return "0s"
	}
	index := (numGC + 255) % 256
	pauseValue := pauseNs[index]
	// Safe conversion: uint64 to int64
	if pauseValue > 9223372036854775807 { // math.MaxInt64
		return "overflow"
	}
	return time.Duration(int64(pauseValue)).String()
}

// formatLastGCTime safely formats LastGC time to avoid integer overflow
func formatLastGCTime(lastGC uint64) string {
	if lastGC == 0 {
		return "never"
	}
	// Safe conversion: uint64 to int64
	if lastGC > 9223372036854775807 { // math.MaxInt64
		return "overflow"
	}
	return time.Unix(0, int64(lastGC)).Format(time.RFC3339)
}
