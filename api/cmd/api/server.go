package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"decorebator.com/internal/app"
	"decorebator.com/internal/common"
	decorebator "decorebator.com/internal/http"
	"github.com/gin-gonic/gin"
)

const productionShutdownTimeout = 5 * time.Second

type serverShutdowner interface {
	Shutdown(context.Context) error
	Close() error
}

func main() {
	if err := run(); err != nil {
		log.Printf("API server stopped with an error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	if os.Getenv("ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	if pgbouncerURL := os.Getenv("PGBOUNCER_DATABASE_URL"); pgbouncerURL != "" {
		os.Setenv("DATABASE_URL", pgbouncerURL)
	}

	// Initialize Context with all services
	appCtx, err := app.NewContext().
		WithDatabase(common.GetDBConnection()).
		WithEnvironment(os.Getenv("ENV")).
		Build()
	if err != nil {
		return fmt.Errorf("initialize application context: %w", err)
	}
	defer appCtx.Close()

	var srv = &http.Server{
		Addr:              ":" + os.Getenv("PORT"),
		Handler:           decorebator.SetupRoutes(appCtx),
		ReadHeaderTimeout: 10 * time.Second, // Fix G112: Prevent Slowloris attacks
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Run server in a goroutine so that it doesn't block. Report startup/runtime
	// failures back to the main goroutine so deferred application cleanup still
	// runs; log.Fatal in this goroutine would terminate the process immediately.
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- srv.ListenAndServe()
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscanll.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	select {
	case <-quit:
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		return nil
	}
	log.Println("Shutdown Server ...")

	if err := shutdownHTTPServer(srv, os.Getenv("ENV"), productionShutdownTimeout); err != nil {
		return fmt.Errorf("stop HTTP server: %w", err)
	}

	log.Println("Server exiting")
	return nil
}

func shutdownHTTPServer(server serverShutdowner, environment string, timeout time.Duration) error {
	ctx := context.Background()
	var cancel context.CancelFunc

	if environment == "production" {
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	if err := server.Shutdown(ctx); err != nil {
		if closeErr := server.Close(); closeErr != nil {
			return errors.Join(
				fmt.Errorf("graceful shutdown: %w", err),
				fmt.Errorf("force close: %w", closeErr),
			)
		}
		return fmt.Errorf("graceful shutdown: %w", err)
	}

	return nil
}
