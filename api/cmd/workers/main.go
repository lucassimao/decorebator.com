package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"decorebator.com/internal/app"
	"decorebator.com/internal/common"
	"decorebator.com/internal/service"
	"github.com/getsentry/sentry-go"
)

func main() {
	// Initialize Sentry first, before any logging
	if err := common.InitSentry(); err != nil {
		common.Logger.Error("Failed to initialize Sentry for workers", "error", err)
	}
	defer func() {
		// Flush any pending Sentry events before shutdown
		if sentry.CurrentHub().Client() != nil {
			sentry.Flush(2 * time.Second)
		}
	}()

	// Initialize database connection
	db := common.GetDBConnection()
	defer common.CloseDBConnection()

	// Create AppContext with all services
	appContext, err := app.NewContext().
		WithDatabase(db).
		WithEnvironment(os.Getenv("ENV")).
		Build()

	if err != nil {
		common.Logger.Error("failed to create app context", "error", err)
		panic(err)
	}
	defer appContext.Close()

	// Create River client for workers using individual services
	riverClient, err := service.NewWorkerRiverClient(
		appContext.Database,
		appContext.DefinitionService,
		appContext.DefinitionImageService,
		appContext.WordService,
		appContext.UserService,
		appContext.LeitnerSystemStrategy,
		appContext.JobService,
		appContext.RevenueCatService,
		appContext.SubscriptionService,
		appContext.MailService,
	)

	if err != nil {
		common.Logger.Error("failed to create river client", "error", err)
		panic(err)
	}

	if err := riverClient.Start(context.Background()); err != nil {
		common.Logger.Error("failed to start river client", "error", err)
		panic(err)
	}

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscanll.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	common.Logger.Debug("Starting backgroundjob shutdown")

	// Gracefully stop River client with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := riverClient.Stop(shutdownCtx); err != nil {
		common.Logger.Error("failed to stop river client", "error", err)
		// Don't panic here - we still want to flush Sentry events
	}

	// Flush any remaining Sentry events before exit
	if sentry.CurrentHub().Client() != nil {
		common.Logger.Debug("Flushing Sentry events")
		if !sentry.Flush(5 * time.Second) {
			common.Logger.Warn("Some Sentry events may not have been sent")
		}
	}

	common.Logger.Debug("backgroundjob shutdown finished")
}
