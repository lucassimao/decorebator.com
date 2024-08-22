package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"decorebator.com/internal/common"
	decorebator "decorebator.com/internal/http"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var origin string

		if os.Getenv("ENV") == "production" {
			origin = "https://decorebator.com"
		} else {
			origin = c.Request.Header.Get("Origin")
		}
		c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Type, Content-Length, Authorization, Cookie")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Cookie")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func main() {

	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	var WordRoutes = decorebator.WordRoutes{}
	var WorkerRoutes = decorebator.WorkerRoutes{}
	var WordlistRoutes = decorebator.WordlistsRoutes{}
	var UserRoutes = decorebator.UserRoutes{}
	var QuizRoutes = decorebator.QuizRoutes{}

	router := gin.Default()

	router.Use(CORSMiddleware())

	// Routes without authentication
	router.POST("/users", UserRoutes.SignUp)
	router.GET("/logout", UserRoutes.Logout)
	router.POST("/login", UserRoutes.Login)

	// Routes with authentication
	authenticatedRoutes := router.Group("/")
	authenticatedRoutes.Use(decorebator.Authenticate)
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

	}

	workerRoutes := router.Group("/static/workers")
	workerRoutes.Use(decorebator.AuthenticateStatic)
	{
		workerRoutes.POST("/imageGenerator/:definitionId", WorkerRoutes.GenerateNewImage)
		workerRoutes.POST("/textToAudio/:wordId", WorkerRoutes.GenerateNewAudio)
		workerRoutes.POST("/retry/:jobId", WorkerRoutes.TriggerJob)
	}

	srv := &http.Server{
		Addr:    ":" + os.Getenv("PORT"),
		Handler: router,
	}

	// Run server in a goroutine so that it doesn't block
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	// kill (no param) default send syscanll.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	if os.Getenv("ENV") == "production" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// catching ctx.Done(). timeout of 5 seconds.
		select {
		case <-ctx.Done():
			log.Println("timeout of 5 seconds.")
		}
	} else {
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Fatal("Server Shutdown:", err)
		}
	}

	log.Println("Server exiting")
	defer common.CloseDBConnection()
}
