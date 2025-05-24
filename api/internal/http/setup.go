package http

import (
	"github.com/gin-gonic/gin"
)

func SetupRoutes() *gin.Engine {

	var WordRoutes = WordRoutes{}
	var WorkerRoutes = WorkerRoutes{}
	var WordlistRoutes = WordlistsRoutes{}
	var UserRoutes = UserRoutes{}
	var QuizRoutes = QuizRoutes{}
	var ErrorReportsRoutes = ErrorReportRoutes{}

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

	}

	// Routes with authentication
	authenticatedRoutes := router.Group("/")
	authenticatedRoutes.Use(Authenticate)
	{
		authenticatedRoutes.GET("/wordlists", WordlistRoutes.GetAll)
		authenticatedRoutes.POST("/wordlists", WordlistRoutes.Create)
		authenticatedRoutes.GET("/wordlists/:wordlistId", WordlistRoutes.GetById)
		authenticatedRoutes.PUT("/wordlists/:wordlistId", WordlistRoutes.Update)
		authenticatedRoutes.DELETE("/wordlists/:wordlistId", WordlistRoutes.Delete)
		authenticatedRoutes.GET("/wordlists/:wordlistId/words", WordRoutes.GetAll)
		authenticatedRoutes.DELETE("/wordlists/:wordlistId/words/:wordId", WordRoutes.Delete)
		authenticatedRoutes.PUT("/wordlists/:wordlistId/words/:wordId", WordRoutes.Update)
		authenticatedRoutes.POST("/wordlists/:wordlistId/words", WordRoutes.Create)
		authenticatedRoutes.POST("/wordlists/:wordlistId/quizzes", QuizRoutes.Create)
		authenticatedRoutes.PATCH("/wordlists/:wordlistId/quizzes", QuizRoutes.Save)
		authenticatedRoutes.POST("/errorReports", ErrorReportsRoutes.Create)
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
