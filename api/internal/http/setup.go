package http

import (
	"time"

	"decorebator.com/internal/app"
	"decorebator.com/internal/common"
	"decorebator.com/internal/model"
	"decorebator.com/internal/repository"
	"decorebator.com/internal/service"
	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

func init() {
	// Initialize Sentry using the common initialization function
	if err := common.InitSentry(); err != nil {
		common.Logger.Error("Failed to initialize Sentry in http package", "error", err)
	} else {
		sentry.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetTag("service", "api")
		})
	}
}

// SetupRoutes creates a Gin engine with routes using Context for dependency injection
func SetupRoutes(appCtx *app.Context) *gin.Engine {
	if appCtx == nil {
		panic("Context cannot be nil")
	}

	// Create repository instances
	subRepo := repository.NewSubscriptionRepository(appCtx.Database)
	userRepo := &repository.UserRepository{Db: appCtx.Database}
	authenticate := Authenticate(appCtx.AuthSessions, userRepo)
	pushTokenRepo := &repository.PushTokenRepository{Db: appCtx.Database}

	// Initialize route handlers using services from AppContext
	var WordRoutes = NewWordRoutes(appCtx.WordService, appCtx.WordlistService, appCtx.DefinitionService)
	var WorkerRoutes = NewWorkerRoutes(appCtx.DefinitionService, appCtx.JobService)
	var WordlistRoutes = NewWordlistsRoutes(appCtx.WordlistService, appCtx.WordService, appCtx.DefinitionService)
	var UserRoutes = NewUserRoutes(
		appCtx.UserService, appCtx.AuthSessions, appCtx.AuthRateLimiter, appCtx.JobService,
	)
	// Use Leitner strategy from AppContext
	var quizRoutes = NewQuizRoutes(appCtx.LeitnerSystemStrategy, userRepo)
	var ErrorReportsRoutes = NewErrorReportRoutes(appCtx.ErrorReportService)
	var PushNotificationRoutes = NewPushNotificationRoutes(pushTokenRepo)

	router := gin.New()
	if err := router.SetTrustedProxies(appCtx.TrustedProxyCIDRs); err != nil {
		panic("invalid trusted proxy configuration: " + err.Error())
	}

	// Sentry middlewares (includes sentrygin + context capture) - completely self-contained
	router.Use(SentryMiddlewares()...)
	router.Use(gin.Logger())
	router.Use(ErrorMiddleware())
	router.Use(CORSMiddleware())
	RegisterHealthRoutes(router, appCtx.Database)
	storeWebhookMetrics := NewStoreWebhookMetrics()
	RegisterStoreWebhookRoutes(router, appCtx, storeWebhookMetrics)
	RegisterMobileIAPRoutes(router, appCtx, appCtx.StoreIAPRequestLimiter, authenticate)

	// Routes without authentication
	{
		router.POST("/users", RateLimitAuthSource(appCtx.AuthRateLimiter, service.AuthLimitSignup), UserRoutes.SignUp)
		router.POST("/logout", UserRoutes.Logout)
		router.POST("/session/refresh", UserRoutes.Refresh)
		router.POST("/login", RateLimitAuthSource(appCtx.AuthRateLimiter, service.AuthLimitLogin), UserRoutes.Login)
		router.PATCH("/password/reset", RateLimitAuthSource(appCtx.AuthRateLimiter, service.AuthLimitResetConsume), UserRoutes.ResetPassword)
		router.POST("/password/send-reset-email", RateLimitAuthSource(appCtx.AuthRateLimiter, service.AuthLimitResetRequest), UserRoutes.SendResetPasswordEmail)

		RegisterLegacyProviderPublicRoutes(router, appCtx)

		// Deprecated demo quiz endpoint removed
	}

	// Routes with authentication
	authenticatedRoutes := router.Group("/")
	authenticatedRoutes.Use(authenticate, ResolveEffectiveSubscription(appCtx.EffectiveAccessService), SentryUserContextMiddleware())
	authenticatedRoutes.Use(TimeoutMiddleware(2 * time.Second))
	{
		authenticatedRoutes.GET("/wordlists", WordlistRoutes.GetAll)
		authenticatedRoutes.POST("/wordlists", CheckSubscriptionLimits(appCtx.SubscriptionService, model.UserActionCreateWordlist), WordlistRoutes.Create)
		authenticatedRoutes.GET("/wordlists/pronunciation-systems", WordlistRoutes.GetPronunciationSystems)
		authenticatedRoutes.GET("/wordlists/:wordlistId", WordlistRoutes.GetByID)
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
		authenticatedRoutes.POST("/errorReports", RateLimitErrorReports(appCtx.Database), ErrorReportsRoutes.Create)
		authenticatedRoutes.GET("/errorReports/status", GetUserErrorReportStatus(appCtx.Database))

		RegisterAnalyticsRoutes(authenticatedRoutes, appCtx.WordlistService, appCtx.Database)
		RegisterRealtimeTelemetryRoutes(authenticatedRoutes, appCtx.RealtimeTelemetryService)

		// Subscription routes
		authenticatedRoutes.GET("/subscription/status", GetSubscriptionStatus(subRepo))
		authenticatedRoutes.GET("/subscription/history", GetSubscriptionHistory(subRepo))
		RegisterLegacyProviderAuthenticatedRoutes(authenticatedRoutes, appCtx)

		// User profile routes
		authenticatedRoutes.GET("/users", UserRoutes.GetProfile)
		authenticatedRoutes.PATCH("/users", UserRoutes.UpdateProfile)
		authenticatedRoutes.DELETE("/users", UserRoutes.DeleteProfile)

		// Push notification routes
		authenticatedRoutes.POST("/push/register", PushNotificationRoutes.Register)
		authenticatedRoutes.POST("/push/unregister", PushNotificationRoutes.Unregister)
	}

	// OpenAI Realtime token creation needs a wider provider-call budget than
	// ordinary authenticated database routes while retaining the same guards.
	chatRoutes := router.Group("/")
	chatRoutes.Use(authenticate, ResolveEffectiveSubscription(appCtx.EffectiveAccessService), SentryUserContextMiddleware())
	chatRoutes.Use(TimeoutMiddleware(15 * time.Second))
	chatRoutes.POST(
		"/wordlists/:wordlistId/chat/session",
		CheckSubscriptionLimits(appCtx.SubscriptionService, model.UserActionChatSession),
		WordlistRoutes.CreateChatSession,
	)

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
		adminRoutes.GET("/store-webhooks/metrics", GetStoreWebhookMetrics(storeWebhookMetrics))
	}

	// Static-auth protected endpoints (keep original paths)
	return router
}

// RegisterLegacyProviderPublicRoutes keeps the rollback surface explicit and
// makes legacy provider ingress and redirects unreachable at IAP-only cutover.
func RegisterLegacyProviderPublicRoutes(router *gin.Engine, appCtx *app.Context) {
	if !appCtx.LegacyProviderSurfaceEnabled {
		return
	}
	router.POST("/webhook/stripe", HandleStripeWebhook(appCtx.SubscriptionService, appCtx.JobService))
	router.POST("/webhook/revenuecat", HandleRevenueCatWebhook(
		appCtx.JobService, appCtx.LegacyRevenueCatWebhookAuth,
	))
	router.GET("/subscription/checkout-redirect", CheckoutRedirect())
}

func RegisterLegacyProviderAuthenticatedRoutes(routes *gin.RouterGroup, appCtx *app.Context) {
	if !appCtx.LegacyProviderSurfaceEnabled {
		return
	}
	routes.POST("/subscription/checkout-session", CreateCheckoutSession(appCtx.SubscriptionService))
	routes.POST("/subscription/revenuecat/restore", RestorePurchases(appCtx.RevenueCatService))
}

func RegisterStoreWebhookRoutes(
	router *gin.Engine,
	appCtx *app.Context,
	observer StoreWebhookObserver,
) {
	if !appCtx.StoreIAPEnabled {
		return
	}
	if appCtx.AppleNotificationIngestor == nil || appCtx.GoogleRTDNIngestor == nil {
		panic("store IAP is enabled without initialized notification receivers")
	}
	router.POST(
		"/webhook/app-store",
		TimeoutMiddleware(28*time.Second),
		HandleAppleStoreNotification(appCtx.AppleNotificationIngestor, observer),
	)
	router.POST(
		"/webhook/google-play/rtdn",
		TimeoutMiddleware(28*time.Second),
		HandleGooglePlayRTDN(appCtx.GoogleRTDNIngestor, observer),
	)
}
