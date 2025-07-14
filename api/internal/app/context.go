package app

import (
	"errors"
	"fmt"
	"log"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/mail"
	"decorebator.com/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

// Context holds all application services and dependencies
type Context struct {
	// Core dependencies
	Database   *pgxpool.Pool
	JobService service.JobService

	// Services
	WordService            *service.WordService
	WordlistService        *service.WordlistService
	UserService            *service.UserService
	DefinitionService      *service.DefinitionService
	DefinitionImageService *service.DefinitionImageService
	SubscriptionService    *service.SubscriptionService
	RevenueCatService      service.RevenueCatService
	ModerationService      service.ModerationService
	LeitnerTrackingService *service.LeitnerTrackingService
	LeitnerSystemStrategy  *service.LeitnerSystemStrategy
	ErrorReportService     *service.ErrorReportService
	AnalyticsService       service.AnalyticsServiceInterface
	MailService            *mail.MailService

	// Configuration
	Environment string
}

// ContextBuilder provides a fluent API for building Context
type ContextBuilder struct {
	context                  *Context
	errors                   []error
	revenueCatServiceFactory func(db *pgxpool.Pool) service.RevenueCatService
}

// NewContext creates a new Context builder
func NewContext() *ContextBuilder {
	return &ContextBuilder{
		context: &Context{},
		errors:  make([]error, 0),
	}
}

// WithDatabase sets the database connection
func (b *ContextBuilder) WithDatabase(db *pgxpool.Pool) *ContextBuilder {
	if db == nil {
		b.errors = append(b.errors, errors.New("database connection cannot be nil"))
		return b
	}
	b.context.Database = db
	return b
}

// WithJobService sets the JobService for background job operations
func (b *ContextBuilder) WithJobService(jobService service.JobService) *ContextBuilder {
	if jobService == nil {
		b.errors = append(b.errors, errors.New("job service cannot be nil"))
		return b
	}
	b.context.JobService = jobService
	return b
}

// WithEnvironment sets the application environment
func (b *ContextBuilder) WithEnvironment(env string) *ContextBuilder {
	b.context.Environment = env
	return b
}

// WithModerationService sets a custom moderation service
func (b *ContextBuilder) WithModerationService(moderationService service.ModerationService) *ContextBuilder {
	b.context.ModerationService = moderationService
	return b
}

// WithRevenueCatService sets a custom RevenueCat service
func (b *ContextBuilder) WithRevenueCatService(revenueCatService service.RevenueCatService) *ContextBuilder {
	b.context.RevenueCatService = revenueCatService
	return b
}

// WithWordService sets a custom word service
func (b *ContextBuilder) WithWordService(wordService *service.WordService) *ContextBuilder {
	b.context.WordService = wordService
	return b
}

// WithWordlistService sets a custom wordlist service
func (b *ContextBuilder) WithWordlistService(wordlistService *service.WordlistService) *ContextBuilder {
	b.context.WordlistService = wordlistService
	return b
}

// WithDefinitionService sets a custom definition service
func (b *ContextBuilder) WithDefinitionService(definitionService *service.DefinitionService) *ContextBuilder {
	b.context.DefinitionService = definitionService
	return b
}

// WithDefinitionImageService sets a custom definition image service
func (b *ContextBuilder) WithDefinitionImageService(definitionImageService *service.DefinitionImageService) *ContextBuilder {
	b.context.DefinitionImageService = definitionImageService
	return b
}

// WithUserService sets a custom user service
func (b *ContextBuilder) WithUserService(userService *service.UserService) *ContextBuilder {
	b.context.UserService = userService
	return b
}

// WithSubscriptionService sets a custom subscription service
func (b *ContextBuilder) WithSubscriptionService(subscriptionService *service.SubscriptionService) *ContextBuilder {
	b.context.SubscriptionService = subscriptionService
	return b
}

// WithRevenueCatServiceFunc sets a custom RevenueCat service using a factory function
func (b *ContextBuilder) WithRevenueCatServiceFunc(factory func(db *pgxpool.Pool) service.RevenueCatService) *ContextBuilder {
	b.revenueCatServiceFactory = factory
	return b
}

// WithLeitnerSystemStrategy sets a custom Leitner system strategy
func (b *ContextBuilder) WithLeitnerSystemStrategy(strategy *service.LeitnerSystemStrategy) *ContextBuilder {
	b.context.LeitnerSystemStrategy = strategy
	return b
}

// WithErrorReportService sets a custom error report service
func (b *ContextBuilder) WithErrorReportService(errorReportService *service.ErrorReportService) *ContextBuilder {
	b.context.ErrorReportService = errorReportService
	return b
}

// WithAnalyticsService sets a custom analytics service
func (b *ContextBuilder) WithAnalyticsService(analyticsService service.AnalyticsServiceInterface) *ContextBuilder {
	b.context.AnalyticsService = analyticsService
	return b
}

// Build constructs the Context with all dependencies initialized
func (b *ContextBuilder) Build() (*Context, error) {
	// Check for builder errors first
	if len(b.errors) > 0 {
		return nil, fmt.Errorf("Context validation failed: %v", b.errors)
	}

	// Validate required dependencies
	if b.context.Database == nil {
		return nil, errors.New("database connection is required")
	}

	// Initialize default services if not provided
	if err := b.initializeServices(); err != nil {
		return nil, fmt.Errorf("failed to initialize services: %w", err)
	}

	return b.context, nil
}

// initializeServices creates default service instances
func (b *ContextBuilder) initializeServices() error {
	// Initialize JobService if not provided
	if b.context.JobService == nil {
		// Create minimal River client for job insertion only
		riverClient, err := river.NewClient(riverpgxv5.New(b.context.Database), &river.Config{
			Logger: common.Logger,
		})
		if err != nil {
			return fmt.Errorf("failed to create river client: %w", err)
		}
		b.context.JobService = service.NewJobService(riverClient)
	}

	// Initialize ModerationService if not provided
	if b.context.ModerationService == nil {
		b.context.ModerationService = service.NewOpenAIModerationService()
	}

	// Initialize core services if not provided
	if b.context.DefinitionService == nil {
		b.context.DefinitionService = service.NewDefinitionService(b.context.Database)
	}
	if b.context.DefinitionImageService == nil {
		b.context.DefinitionImageService = service.NewDefinitionImageService(b.context.Database)
	}

	// Initialize LeitnerTrackingService if not provided
	if b.context.LeitnerTrackingService == nil {
		b.context.LeitnerTrackingService = service.NewLeitnerTrackingService(b.context.Database)
	}

	if b.context.WordService == nil {
		b.context.WordService = service.NewWordService(b.context.Database, b.context.DefinitionService, b.context.ModerationService, b.context.JobService, b.context.LeitnerTrackingService)
	}
	if b.context.WordlistService == nil {
		b.context.WordlistService = service.NewWordlistService(b.context.Database, b.context.ModerationService)
	}

	// Initialize MailService early since other services depend on it
	if b.context.MailService == nil {
		b.context.MailService = mail.NewMailService(b.context.Database)
	}

	if b.context.SubscriptionService == nil {
		b.context.SubscriptionService = service.NewSubscriptionService(b.context.Database, b.context.MailService)
	}

	// Initialize RevenueCatService if not provided
	if b.context.RevenueCatService == nil {
		if b.revenueCatServiceFactory != nil {
			b.context.RevenueCatService = b.revenueCatServiceFactory(b.context.Database)
		} else {
			apiClient := service.NewRevenueCatAPIClient()
			b.context.RevenueCatService = service.NewRevenueCatService(b.context.Database, apiClient)
		}
	}

	// Initialize AnalyticsService if not provided
	if b.context.AnalyticsService == nil {
		// Create a base analytics service without user/wordlist specifics for general use
		analyticsService, err := service.NewAnalyticsService(b.context.Database, service.AnalyticsConfig{
			UserID:     0, // Will be set by individual handlers
			WordlistID: 0, // Will be set by individual handlers
			UseCache:   true,
			CacheTTL:   1 * time.Minute,
		})
		if err != nil {
			return fmt.Errorf("failed to create analytics service: %w", err)
		}
		b.context.AnalyticsService = analyticsService
	}

	// Initialize LeitnerSystemStrategy if not provided
	if b.context.LeitnerSystemStrategy == nil {
		// Create analytics writer locally for LeitnerSystemStrategy (non-cached)
		analyticsService, err := service.NewAnalyticsService(b.context.Database, service.AnalyticsConfig{
			UserID:     0,     // Not used for stateless operations
			WordlistID: 0,     // Not used for stateless operations
			UseCache:   false, // No caching for write operations
			CacheTTL:   0,
		})
		if err != nil {
			return fmt.Errorf("failed to create analytics writer: %w", err)
		}

		// Cast to interface pointer
		var analyticsWriter service.LeitnerAnalyticsWriter = analyticsService

		// Create strategy with analytics writer as internal dependency
		b.context.LeitnerSystemStrategy = service.NewLeitnerSystemStrategy(
			b.context.Database,
			b.context.WordService,
			b.context.DefinitionService,
			&analyticsWriter,
			b.context.LeitnerTrackingService,
		)
	}

	// Initialize ErrorReportService if not provided
	if b.context.ErrorReportService == nil {
		b.context.ErrorReportService = service.NewErrorReportService(
			b.context.Database,
			b.context.DefinitionService,
			b.context.WordService,
			b.context.LeitnerTrackingService,
			b.context.JobService,
		)
	}

	// Initialize UserService after ErrorReportService
	if b.context.UserService == nil {
		b.context.UserService = service.NewUserService(b.context.Database, b.context.SubscriptionService, b.context.ErrorReportService)
	}

	return nil
}

// Close gracefully shuts down all services and connections
func (ctx *Context) Close() {
	log.Println("Shutting down Context...")

	// Close database connection
	if ctx.Database != nil {
		ctx.Database.Close()
		log.Println("Database connection closed")
	}

	log.Println("Context shutdown complete")
}

// GetDatabase returns the database connection
func (ctx *Context) GetDatabase() *pgxpool.Pool {
	return ctx.Database
}

// GetJobService returns the JobService
func (ctx *Context) GetJobService() service.JobService {
	return ctx.JobService
}
