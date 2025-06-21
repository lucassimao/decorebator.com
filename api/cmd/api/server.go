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

func main() {

	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	var srv = &http.Server{
		Addr:              ":" + os.Getenv("PORT"),
		Handler:           decorebator.SetupRoutes(nil),
		ReadHeaderTimeout: 10 * time.Second, // Fix G112: Prevent Slowloris attacks
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	defer common.CloseDBConnection()

	// Run server in a goroutine so that it doesn't block
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
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
}
