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
	"decorebator.com/internal/users"
	"decorebator.com/internal/wordlists"
	"decorebator.com/internal/words"
	"github.com/gin-gonic/gin"
)

func main() {

	router := gin.Default()
	// router.Use(common.ValidateIDParamMidleware())

	router.GET("/wordlists", wordlists.Handlers.GetAll)
	router.POST("/wordlists", wordlists.Handlers.Create)
	router.GET("/wordlists/:wordlistId", wordlists.Handlers.GetById)
	router.PUT("/wordlists/:wordlistId", wordlists.Handlers.Update)
	router.DELETE("/wordlists/:wordlistId", wordlists.Handlers.Delete)
	router.GET("/wordlists/:wordlistId/words")
	router.GET("/wordlists/:wordlistId/words/:wordId", words.Handlers.GetAll)
	router.DELETE("/wordlists/:wordlistId/words/:wordId", words.Handlers.Delete)
	router.PATCH("/wordlists/:wordlistId/words/:wordId", words.Handlers.Update)
	router.POST("/wordlists/:wordlistId/words", words.Handlers.Create)

	router.POST("/users", users.Handlers.SignUp)
	router.POST("/login", users.Handlers.Login)

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

	if os.Getenv("PORT") == "production" {
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
