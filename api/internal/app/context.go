package app

import (
	"errors"
	"fmt"
	"log"

	"decorebator.com/internal/service"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
)

// Context holds all application services and dependencies
type Context struct {
	// Core dependencies
	Database    *pgxpool.Pool
	RiverClient *river.Client[pgx.Tx]

	// Services
	WordService            *service.WordService
	WordlistService        *service.WordlistService
	UserService            *service.UserService
	DefinitionService      *service.DefinitionService
	DefinitionImageService *service.DefinitionImageService
	SubscriptionService    *service.SubscriptionService
	RevenueCatService      service.RevenueCatService
	ModerationService      service.ModerationService
	LeitnerSystemStrategy  *service.LeitnerSystemStrategy
	ErrorReportService     *service.ErrorReportService

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

// WithRiverClient sets the River client for background jobs
func (b *ContextBuilder) WithRiverClient(client *river.Client[pgx.Tx]) *ContextBuilder {
	if client == nil {
		b.errors = append(b.errors, errors.New("river client cannot be nil"))
		return b
	}
	b.context.RiverClient = client
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
	// Initialize RiverClient if not provided
	if b.context.RiverClient == nil {
		riverClient, err := service.GetRiverClient()
		if err != nil {
			return fmt.Errorf("failed to create river client: %w", err)
		}
		b.context.RiverClient = riverClient
	}

	// Initialize ModerationService if not provided
	if b.context.ModerationService == nil {
		b.context.ModerationService = service.NewOpenAIModerationService()
	}

	// Initialize core services if not provided
	if b.context.UserService == nil {
		b.context.UserService = service.NewUserService(b.context.Database)
	}
	if b.context.DefinitionService == nil {
		b.context.DefinitionService = service.NewDefinitionService(b.context.Database)
	}
	if b.context.DefinitionImageService == nil {
		b.context.DefinitionImageService = service.NewDefinitionImageService(b.context.Database)
	}
	if b.context.WordService == nil {
		b.context.WordService = service.NewWordService(b.context.Database, b.context.ModerationService)
	}
	if b.context.WordlistService == nil {
		b.context.WordlistService = service.NewWordlistService(b.context.Database, b.context.ModerationService)
	}
	if b.context.SubscriptionService == nil {
		b.context.SubscriptionService = service.NewSubscriptionService(b.context.Database)
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

	// Initialize LeitnerSystemStrategy if not provided
	if b.context.LeitnerSystemStrategy == nil {
		b.context.LeitnerSystemStrategy = service.NewLeitnerSystemStrategy(
			b.context.WordService,
			b.context.DefinitionService,
		)
	}

	// Initialize ErrorReportService if not provided
	if b.context.ErrorReportService == nil {
		b.context.ErrorReportService = service.NewErrorReportService(
			b.context.Database,
			b.context.DefinitionService,
			b.context.LeitnerSystemStrategy,
		)
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

// GetRiverClient returns the River client
func (ctx *Context) GetRiverClient() *river.Client[pgx.Tx] {
	return ctx.RiverClient
}
