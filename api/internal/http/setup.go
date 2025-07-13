package http

import (
	"fmt"
	"log"
	"os"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/repository"
	"decorebator.com/internal/service"
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
)

// Config holds optional configuration for setting up routes
type Config struct {
	WordService            *service.WordService
	WordlistService        *service.WordlistService
	DefinitionService      *service.DefinitionService
	DefinitionImageService *service.DefinitionImageService
	ModerationService      service.ModerationService
	Database               *pgxpool.Pool
	riverClient            *river.Client[pgx.Tx]
	RevenueCatService      service.RevenueCatService
}

func init() {
	sentryDsn, exists := os.LookupEnv("SENTRY_DSN")
	if os.Getenv("ENV") == productionEnv && !exists {
		panic("SENTRY_DSN not found")
	}

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:            sentryDsn,
		Debug:          false, // Set to false in production to reduce overhead
		SendDefaultPII: true,
		// Performance optimizations
		SampleRate:       1.0, // Capture 100% of errors
		TracesSampleRate: 0.1, // Only capture 10% of performance traces
		// Buffering settings
		MaxBreadcrumbs: 50,
		// The SDK handles async sending internally
		AttachStacktrace: true,
	}); err != nil {
		common.Logger.Error("Sentry initialization failed", "error", err)
	}

}

// applyDefaults fills in missing configuration with default values
func applyDefaults(config *Config) (*Config, error) {
	// Start with empty config if nil
	if config == nil {
		config = &Config{}
	}

	// Set riverClient if not provided
	if config.riverClient == nil {
		riverClient, err := service.GetRiverClient()
		if err != nil {
			log.Fatalf("failed to create river client: %s\n", err)
		}
		config.riverClient = riverClient
	}

	// Set default database connection if not provided
	if config.Database == nil {
		db := common.GetDBConnection()
		config.Database = db
	}

	// Set default moderation service if not provided
	if config.ModerationService == nil {
		config.ModerationService = service.NewOpenAIModerationService()
	}

	// Set default word service if not provided
	if config.WordService == nil {
		config.WordService = service.NewWordService(config.Database, config.ModerationService)
	}

	// Set default wordlist service if not provided
	if config.WordlistService == nil {
		config.WordlistService = service.NewWordlistService(config.Database, config.ModerationService)
	}

	// Set default DefinitionService if not provided
	if config.DefinitionService == nil {
		config.DefinitionService = service.NewDefinitionService(config.Database)
	}

	// Set default DefinitionImageService if not provided
	if config.DefinitionImageService == nil {
		config.DefinitionImageService = service.NewDefinitionImageService(config.Database)
	}

	// Set default RevenueCat service if not provided
	if config.RevenueCatService == nil {
		apiClient := service.NewRevenueCatAPIClient()
		config.RevenueCatService = service.NewRevenueCatService(config.Database, apiClient)
	}

	return config, nil
}

func SetupRoutes(config *Config) *gin.Engine {
	// Apply defaults to configuration
	config, err := applyDefaults(config)
	if err != nil {
		panic(fmt.Sprintf("Failed to apply default configuration: %v", err))
	}

	subService := service.NewSubscriptionService(config.Database)
	subRepo := repository.NewSubscriptionRepository(config.Database)
	rcService := config.RevenueCatService

	// Initialize route handlers with dependency injection
	userService := service.NewUserService(config.Database)
	var WordRoutes = NewWordRoutes(config.WordService, config.DefinitionService)
	var WorkerRoutes = NewWorkerRoutes(config.DefinitionService)
	var WordlistRoutes = NewWordlistsRoutes(config.WordlistService, config.WordService)
	var UserRoutes = NewUserRoutes(userService)

	// Create Leitner strategy with proper dependency injection
	strategy := service.NewLeitnerSystemStrategy(config.WordService, config.DefinitionService)
	var quizRoutes = NewQuizRoutes(strategy)

	var ErrorReportsRoutes = ErrorReportRoutes{}
	var demoRoutes = DemoRoutes{}

	router := gin.New()
	// Sentry middlewares (includes sentrygin + context capture) - completely self-contained
	router.Use(SentryMiddlewares()...)
	router.Use(gin.Logger())
	router.Use(TimeoutMiddleware(2 * time.Second))
	router.Use(ErrorMiddleware())
	router.Use(CORSMiddleware())

	// Routes without authentication
	{
		router.POST("/users", UserRoutes.SignUp)
		router.GET("/logout", UserRoutes.Logout)
		router.POST("/login", UserRoutes.Login)
		router.PATCH("/password/reset", UserRoutes.ResetPassword)
		router.POST("/password/send-reset-email", UserRoutes.SendResetPasswordEmail)

		// Stripe webhook endpoint
		router.POST("/webhook/stripe", HandleStripeWebhook(subService, config.riverClient))

		// RevenueCat webhook endpoint
		router.POST("/webhook/revenuecat", HandleRevenueCatWebhook(config.riverClient))

		// Redirect to local expo scheme
		router.GET("/subscription/checkout-redirect", CheckoutRedirect())

		// Demo quiz endpoint for web landing page (static auth)
		router.GET("/quiz/demo", AuthenticateStatic, demoRoutes.GetDemoQuizzes)
	}

	// Routes with authentication
	authenticatedRoutes := router.Group("/")
	authenticatedRoutes.Use(Authenticate, SentryUserContextMiddleware())
	{
		authenticatedRoutes.GET("/wordlists", WordlistRoutes.GetAll)
		authenticatedRoutes.POST("/wordlists", CheckSubscriptionLimits(subService, "create_wordlist"), WordlistRoutes.Create)
		authenticatedRoutes.GET("/wordlists/pronunciation-systems", WordlistRoutes.GetPronunciationSystems)
		authenticatedRoutes.GET("/wordlists/:wordlistId", WordlistRoutes.GetById)
		authenticatedRoutes.PUT("/wordlists/:wordlistId", WordlistRoutes.Update)
		authenticatedRoutes.DELETE("/wordlists/:wordlistId", WordlistRoutes.Delete)
		authenticatedRoutes.GET("/wordlists/:wordlistId/processing-status", WordlistRoutes.GetProcessingStatus)
		authenticatedRoutes.GET("/wordlists/:wordlistId/words", WordRoutes.GetAll)
		authenticatedRoutes.DELETE("/wordlists/:wordlistId/words/:wordId", WordRoutes.Delete)
		authenticatedRoutes.PUT("/wordlists/:wordlistId/words/:wordId", WordRoutes.Update)
		authenticatedRoutes.GET("/wordlists/:wordlistId/words/:wordId/definitions", WordRoutes.GetDefinitions)
		authenticatedRoutes.POST("/wordlists/:wordlistId/words", CheckSubscriptionLimits(subService, "add_word"), WordRoutes.Create)
		authenticatedRoutes.POST("/wordlists/:wordlistId/quizzes", quizRoutes.Create)
		authenticatedRoutes.PATCH("/wordlists/:wordlistId/quizzes", quizRoutes.Save)
		authenticatedRoutes.POST("/errorReports", RateLimitErrorReports(), ErrorReportsRoutes.Create)
		authenticatedRoutes.GET("/errorReports/status", GetUserErrorReportStatus())

		RegisterAnalyticsRoutes(authenticatedRoutes)

		// Subscription routes
		authenticatedRoutes.POST("/subscription/checkout-session", CreateCheckoutSession(subService))
		authenticatedRoutes.GET("/subscription/status", GetSubscriptionStatus(subRepo))
		authenticatedRoutes.POST("/subscription/cancel", CancelSubscription(subService))
		authenticatedRoutes.GET("/subscription/history", GetSubscriptionHistory(subRepo))
		// RevenueCat routes
		authenticatedRoutes.POST("/subscription/revenuecat/restore", RestorePurchases(rcService))

		// User profile routes
		authenticatedRoutes.GET("/users", UserRoutes.GetProfile)
		authenticatedRoutes.PATCH("/users", UserRoutes.UpdateProfile)
		authenticatedRoutes.DELETE("/users", UserRoutes.DeleteProfile)

	}

	workerRoutes := router.Group("/static/workers")
	workerRoutes.Use(AuthenticateStatic)
	{
		workerRoutes.POST("/imageGenerator/:definitionId", WorkerRoutes.GenerateNewImage)
		workerRoutes.POST("/textToAudio/:wordId", WorkerRoutes.GenerateNewAudio)
		workerRoutes.POST("/definition/:wordId", WorkerRoutes.GenerateNewDefinition)
		workerRoutes.POST("/retry/:jobId", WorkerRoutes.TriggerJob)
	}

	// Admin routes with static authentication
	adminRoutes := router.Group("/static/admin")
	adminRoutes.Use(AuthenticateStatic)
	{
		adminRoutes.GET("/errorReports/stats", GetErrorReportStats())
	}

	return router
}
