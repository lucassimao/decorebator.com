package http

import (
	"fmt"
	"os"

	"decorebator.com/internal/common"
	"decorebator.com/internal/repository"
	"decorebator.com/internal/service"
	"github.com/gin-gonic/gin"
)

func SetupRoutes() *gin.Engine {

	var WordRoutes = WordRoutes{}
	var WorkerRoutes = WorkerRoutes{}
	var WordlistRoutes = WordlistsRoutes{}
	var UserRoutes = UserRoutes{}
	var QuizRoutes = QuizRoutes{}
	var ErrorReportsRoutes = ErrorReportRoutes{}

	// Initialize subscription service
	db, err := common.GetDBConnection()
	if err != nil {
		common.Logger.Error("Failed to get database connection", "error", err)
		fmt.Fprintf(os.Stderr, "Failed to get database connection: %v\n", err)
		os.Exit(1)
	}

	subService := service.NewSubscriptionService(db)
	subRepo := repository.NewSubscriptionRepository(db)
	userRepo := &repository.UserRepository{Db: db}

	router := gin.New()
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

		// Stripe webhook endpoint (no auth needed)
		router.POST("/webhook/stripe", HandleStripeWebhook(subService))
		// Redirect to local expo scheme
		router.GET("/subscription/checkout-redirect", CheckoutRedirect())
	}

	// Routes with authentication
	authenticatedRoutes := router.Group("/")
	authenticatedRoutes.Use(Authenticate)
	{
		authenticatedRoutes.GET("/wordlists", WordlistRoutes.GetAll)
		authenticatedRoutes.POST("/wordlists", CheckSubscriptionLimits(subService, "create_wordlist"), WordlistRoutes.Create)
		authenticatedRoutes.GET("/wordlists/:wordlistId", WordlistRoutes.GetById)
		authenticatedRoutes.PUT("/wordlists/:wordlistId", WordlistRoutes.Update)
		authenticatedRoutes.DELETE("/wordlists/:wordlistId", WordlistRoutes.Delete)
		authenticatedRoutes.GET("/wordlists/:wordlistId/words", WordRoutes.GetAll)
		authenticatedRoutes.DELETE("/wordlists/:wordlistId/words/:wordId", WordRoutes.Delete)
		authenticatedRoutes.PUT("/wordlists/:wordlistId/words/:wordId", WordRoutes.Update)
		authenticatedRoutes.POST("/wordlists/:wordlistId/words", CheckSubscriptionLimits(subService, "add_word"), WordRoutes.Create)
		authenticatedRoutes.POST("/wordlists/:wordlistId/quizzes", QuizRoutes.Create)
		authenticatedRoutes.PATCH("/wordlists/:wordlistId/quizzes", QuizRoutes.Save)
		authenticatedRoutes.POST("/errorReports", ErrorReportsRoutes.Create)

		// Subscription routes
		authenticatedRoutes.POST("/subscription/checkout-session", CreateCheckoutSession(subService))
		authenticatedRoutes.GET("/subscription/status", GetSubscriptionStatus(subRepo))
		authenticatedRoutes.POST("/subscription/cancel", CancelSubscription(subService))
		authenticatedRoutes.GET("/subscription/history", GetSubscriptionHistory(subRepo))

		// Auth routes
		authenticatedRoutes.POST("/auth/refresh", RefreshToken(userRepo))
	}

	workerRoutes := router.Group("/static/workers")
	workerRoutes.Use(AuthenticateStatic)
	{
		workerRoutes.POST("/imageGenerator/:definitionId", WorkerRoutes.GenerateNewImage)
		workerRoutes.POST("/textToAudio/:wordId", WorkerRoutes.GenerateNewAudio)
		workerRoutes.POST("/retry/:jobId", WorkerRoutes.TriggerJob)
	}

	return router
}
