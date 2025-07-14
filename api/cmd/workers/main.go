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
)

func main() {
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

	if os.Getenv("ENV") == "production" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// catching ctx.Done(). timeout of 5 seconds.
		select {
		case <-ctx.Done():
			common.Logger.Debug("timeout of 5 seconds.")
		}
	} else {
		if err := riverClient.Stop(context.Background()); err != nil {
			common.Logger.Error("failed to stop river client", "error", err)
			panic(err)
		}
	}

	common.Logger.Debug("backgroundjob shutdown finished")
}
