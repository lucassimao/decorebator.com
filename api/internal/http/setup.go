package http

import (
	"time"

	"decorebator.com/internal/app"
	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	ddgin "github.com/DataDog/dd-trace-go/contrib/gin-gonic/gin/v2"
	"github.com/gin-gonic/gin"
)

func init() {
	// Initialize Sentry using the common initialization function
	if err := common.InitSentry(); err != nil {
		common.Logger.Error("Failed to initialize Sentry in http package", "error", err)
	}
}

// SetupRoutes creates a Gin engine with routes using Context for dependency injection
func SetupRoutes(appCtx *app.Context) *gin.Engine {
	if appCtx == nil {
		panic("Context cannot be nil")
	}

	// Create repository instances
	subRepo := repository.NewSubscriptionRepository(appCtx.Database)

	// Initialize route handlers using services from AppContext
	var WordRoutes = NewWordRoutes(appCtx.WordService, appCtx.DefinitionService)
	var WorkerRoutes = NewWorkerRoutes(appCtx.DefinitionService, appCtx.JobService)
	var WordlistRoutes = NewWordlistsRoutes(appCtx.WordlistService, appCtx.WordService, appCtx.DefinitionService)
	var UserRoutes = NewUserRoutes(appCtx.UserService, appCtx.MailService)
	var PublicQuizRoutes = NewPublicQuizRoutes(
		repository.NewPublicQuizRepository(appCtx.Database),
		appCtx.WordlistService,
		appCtx.DefinitionService,
		appCtx.Database,
		appCtx.RedisClient,
		appCtx.JobService,
	)

	// Use Leitner strategy from AppContext
	var quizRoutes = NewQuizRoutes(appCtx.LeitnerSystemStrategy)
	var ErrorReportsRoutes = NewErrorReportRoutes(appCtx.ErrorReportService)

	router := gin.New()

	// Datadog middleware (production only)
	if appCtx.DatadogService != nil && appCtx.DatadogService.IsEnabled() {
		router.Use(ddgin.Middleware("decorebator-api"))
		router.Use(DatadogEnrichmentMiddleware()) // Add custom trace enrichment
	}

	// Sentry middlewares (includes sentrygin + context capture) - completely self-contained
	router.Use(SentryMiddlewares()...)
	router.Use(gin.Logger())
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
		router.POST("/webhook/stripe", HandleStripeWebhook(appCtx.SubscriptionService, appCtx.JobService))

		// RevenueCat webhook endpoint
		router.POST("/webhook/revenuecat", HandleRevenueCatWebhook(appCtx.JobService))

		// Redirect to local expo scheme
		router.GET("/subscription/checkout-redirect", CheckoutRedirect())

		// Deprecated demo quiz endpoint removed

		// Public quiz (unauthenticated)
		router.GET("/public-quizzes/:slug", PublicQuizRoutes.GetBySlug)
		router.GET("/public-quizzes/:slug/questions", PublicQuizRoutes.GetQuestionsBySlug)
		router.GET("/public-quizzes/:slug/leaderboard", PublicQuizRoutes.GetLeaderboardBySlug)
		router.POST("/public-quizzes/:slug/attempts", PublicQuizRoutes.RecordAttempt)
	}

	// Routes with authentication
	authenticatedRoutes := router.Group("/")
	authenticatedRoutes.Use(Authenticate, SentryUserContextMiddleware())
	authenticatedRoutes.Use(TimeoutMiddleware(2 * time.Second))
	{
		authenticatedRoutes.GET("/wordlists", WordlistRoutes.GetAll)
		authenticatedRoutes.POST("/wordlists", CheckSubscriptionLimits(appCtx.SubscriptionService, model.UserActionCreateWordlist), WordlistRoutes.Create)
		authenticatedRoutes.GET("/wordlists/pronunciation-systems", WordlistRoutes.GetPronunciationSystems)
		authenticatedRoutes.GET("/wordlists/:wordlistId", WordlistRoutes.GetById)
		authenticatedRoutes.PUT("/wordlists/:wordlistId", WordlistRoutes.Update)
		authenticatedRoutes.DELETE("/wordlists/:wordlistId", WordlistRoutes.Delete)
		authenticatedRoutes.GET("/wordlists/:wordlistId/processing-status", WordlistRoutes.GetProcessingStatus)
		authenticatedRoutes.GET("/wordlists/:wordlistId/words", WordRoutes.GetAll)
		// Batched definitions lookup for multiple word IDs
		authenticatedRoutes.GET("/wordlists/:wordlistId/words/definitions", WordRoutes.GetDefinitionsBatch)
		authenticatedRoutes.DELETE("/wordlists/:wordlistId/words/:wordId", WordRoutes.Delete)
		authenticatedRoutes.PUT("/wordlists/:wordlistId/words/:wordId", WordRoutes.Update)
		authenticatedRoutes.GET("/wordlists/:wordlistId/words/:wordId/definitions", WordRoutes.GetDefinitions)
		authenticatedRoutes.POST("/wordlists/:wordlistId/words", CheckSubscriptionLimits(appCtx.SubscriptionService, model.UserActionAddWord), WordRoutes.Create)
		authenticatedRoutes.POST("/wordlists/:wordlistId/quizzes", quizRoutes.Create)
		authenticatedRoutes.PATCH("/wordlists/:wordlistId/quizzes", quizRoutes.Save)
		// Chat session endpoint - premium only
		authenticatedRoutes.POST("/wordlists/:wordlistId/chat/session", CheckSubscriptionLimits(appCtx.SubscriptionService, model.UserActionChatSession), WordlistRoutes.CreateChatSession)
		// Publish/unpublish endpoints removed; feature disabled in mobile UI
		authenticatedRoutes.POST("/errorReports", RateLimitErrorReports(appCtx.Database), ErrorReportsRoutes.Create)
		authenticatedRoutes.GET("/errorReports/status", GetUserErrorReportStatus(appCtx.Database))

		RegisterAnalyticsRoutes(authenticatedRoutes, appCtx.WordlistService, appCtx.Database)

		// Subscription routes
		authenticatedRoutes.POST("/subscription/checkout-session", CreateCheckoutSession(appCtx.SubscriptionService))
		authenticatedRoutes.GET("/subscription/status", GetSubscriptionStatus(subRepo))
		authenticatedRoutes.GET("/subscription/history", GetSubscriptionHistory(subRepo))
		// RevenueCat routes
		authenticatedRoutes.POST("/subscription/revenuecat/restore", RestorePurchases(appCtx.RevenueCatService))

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
		adminRoutes.GET("/errorReports/stats", GetErrorReportStats(appCtx.Database))

		// System monitoring endpoints
		adminRoutes.GET("/health", HealthCheckHandler(appCtx))
		adminRoutes.GET("/metrics", MetricsHandler(appCtx))
		adminRoutes.GET("/info", SystemInfoHandler())

		// pprof endpoints for performance profiling
		adminRoutes.GET("/debug/pprof/*profile", PprofHandler())
	}

	// Static-auth protected endpoints (keep original paths)
	staticProtected := router.Group("/")
	staticProtected.Use(AuthenticateStatic)
	{
		staticProtected.GET("/public-quizzes", PublicQuizRoutes.ListActive)
	}

	return router
}
