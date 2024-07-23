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
	"decorebator.com/internal/quizzes"
	"decorebator.com/internal/users"
	"decorebator.com/internal/wordlists"
	"decorebator.com/internal/words"
	"decorebator.com/internal/workers"
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

	router := gin.Default()

	router.Use(CORSMiddleware())

	// Routes without authentication
	router.POST("/users", users.Handlers.SignUp)
	router.GET("/logout", users.Handlers.Logout)
	router.POST("/login", users.Handlers.Login)

	// Routes with authentication
	authenticatedRoutes := router.Group("/")
	authenticatedRoutes.Use(users.Handlers.Authenticate)
	{
		authenticatedRoutes.GET("/wordlists", wordlists.Handlers.GetAll)
		authenticatedRoutes.POST("/wordlists", wordlists.Handlers.Create)
		authenticatedRoutes.GET("/wordlists/:wordlistId", wordlists.Handlers.GetById)
		authenticatedRoutes.PUT("/wordlists/:wordlistId", wordlists.Handlers.Update)
		authenticatedRoutes.DELETE("/wordlists/:wordlistId", wordlists.Handlers.Delete)
		authenticatedRoutes.GET("/wordlists/:wordlistId/words", words.Handlers.GetAll)
		authenticatedRoutes.DELETE("/wordlists/:wordlistId/words/:wordId", words.Handlers.Delete)
		authenticatedRoutes.PUT("/wordlists/:wordlistId/words/:wordId", words.Handlers.Update)
		authenticatedRoutes.POST("/wordlists/:wordlistId/words", words.Handlers.Create)
		authenticatedRoutes.POST("/wordlists/:wordlistId/quizzes", quizzes.Handlers.Create)
		authenticatedRoutes.PATCH("/wordlists/:wordlistId/quizzes", quizzes.Handlers.Save)

	}

	workerRoutes := router.Group("/static/workers")
	workerRoutes.Use(users.Handlers.AuthenticateStatic)
	{
		workerRoutes.POST("/imageGenerator/:definitionId", workers.Handlers.GenerateNewImage)
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
