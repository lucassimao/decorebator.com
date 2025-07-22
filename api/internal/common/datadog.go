package common

import (
	"fmt"
	"os"

	"github.com/DataDog/dd-trace-go/v2/ddtrace/tracer"
	"github.com/DataDog/dd-trace-go/v2/profiler"
)

// DatadogService manages Datadog tracer and profiler lifecycle
type DatadogService struct {
	tracerStarted   bool
	profilerStarted bool
}

// NewDatadogService creates a new Datadog service instance
func NewDatadogService() *DatadogService {
	return &DatadogService{
		tracerStarted:   false,
		profilerStarted: false,
	}
}

// Start initializes Datadog tracer and profiler when enabled
func (ds *DatadogService) Start() error {
	// Check if Datadog is enabled
	enabled := os.Getenv("DD_ENABLED")
	if enabled != DDEnabledValue {
		Logger.Info("Datadog disabled - DD_ENABLED not set to true")
		return nil
	}

	// Check required environment variables
	apiKey := os.Getenv("DD_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("DD_API_KEY is required when DD_ENABLED=true")
	}

	serviceName := os.Getenv("DD_SERVICE")
	if serviceName == "" {
		return fmt.Errorf("DD_SERVICE is required for Datadog")
	}

	env := os.Getenv("DD_ENV")
	if env == "" {
		return fmt.Errorf("DD_ENV is required for Datadog")
	}

	version := os.Getenv("DD_VERSION")
	if version == "" {
		return fmt.Errorf("DD_VERSION is required for Datadog")
	}

	// Start tracer with configuration
	if err := ds.startTracer(serviceName, env, version); err != nil {
		Logger.Error("Failed to start Datadog tracer", "error", err)
		return err
	}

	// Start profiler with configuration
	if err := ds.startProfiler(serviceName, env, version); err != nil {
		Logger.Error("Failed to start Datadog profiler", "error", err)
		// Don't fail if profiler doesn't start, tracer is more important
	}

	Logger.Info("Datadog initialized successfully")
	return nil
}

// startTracer initializes the Datadog tracer
func (ds *DatadogService) startTracer(serviceName, env, version string) error {
	options := []tracer.StartOption{
		tracer.WithService(serviceName),
		tracer.WithEnv(env),
		tracer.WithServiceVersion(version),
		tracer.WithRuntimeMetrics(),
		tracer.WithAnalytics(true),
		// Performance optimizations
		tracer.WithSampler(tracer.NewRateSampler(1.0)), // Sample 100% of traces in production
		tracer.WithDebugMode(false),
	}

	// Add site configuration if provided
	if site := os.Getenv("DD_SITE"); site != "" {
		// Note: DD_SITE is typically handled by the Datadog agent configuration
		Logger.Info("Using Datadog site", "site", site)
	}

	tracer.Start(options...) //nolint:errcheck // tracer.Start does not return an error
	ds.tracerStarted = true

	Logger.Info("Datadog tracer started",
		"service", serviceName,
		"env", env,
		"version", version)

	return nil
}

// startProfiler initializes the Datadog profiler
func (ds *DatadogService) startProfiler(serviceName, env, version string) error {
	err := profiler.Start(
		profiler.WithService(serviceName),
		profiler.WithEnv(env),
		profiler.WithVersion(version),
		profiler.WithProfileTypes(
			profiler.CPUProfile,
			profiler.HeapProfile,
			profiler.BlockProfile,
			profiler.MutexProfile,
			profiler.GoroutineProfile,
		),
		// Upload profiles every minute
		profiler.WithPeriod(60),
		profiler.WithTags("component:api", "language:go"),
	)

	if err != nil {
		return err
	}

	ds.profilerStarted = true
	Logger.Info("Datadog profiler started")
	return nil
}

// Stop gracefully shuts down Datadog tracer and profiler
func (ds *DatadogService) Stop() {
	if ds.profilerStarted {
		profiler.Stop()
		Logger.Info("Datadog profiler stopped")
	}

	if ds.tracerStarted {
		tracer.Stop()
		Logger.Info("Datadog tracer stopped")
	}
}

// IsEnabled returns true if Datadog is enabled via DD_ENABLED environment variable
func (ds *DatadogService) IsEnabled() bool {
	return os.Getenv("DD_ENABLED") == DDEnabledValue
}
