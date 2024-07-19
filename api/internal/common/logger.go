package common

import (
	"log/slog"
	"os"
)

var options = &slog.HandlerOptions{Level: slog.LevelDebug}
var handler slog.Handler
var Logger *slog.Logger

func init() {
	if Config.Env == Production {
		handler = slog.NewJSONHandler(os.Stdout, options)
	} else {
		handler = slog.NewTextHandler(os.Stdout, options)
	}
	Logger = slog.New(handler)
}
