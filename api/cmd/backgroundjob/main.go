package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"decorebator.com/internal/common"
	"decorebator.com/internal/workers"
)

func main() {

	riverClient, err := workers.GetRiverClient()

	if err != nil {
		common.Logger.Error("failed to start river client")
		panic(err)
	}

	// Run the client inline. All executed jobs will inherit from ctx:
	if err := riverClient.Start(context.Background()); err != nil {
		common.Logger.Error("failed to start river client")
		panic(err)
	}

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
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
	defer common.CloseDBConnection()
}
