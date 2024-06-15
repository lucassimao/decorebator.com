package common

import (
	"log/slog"
	"os"
)

var options = &slog.HandlerOptions{Level: slog.LevelDebug}
var handler slog.Handler

func init() {
	if Config.env == Development {
		handler = slog.NewJSONHandler(os.Stderr, options)
	} else {
		handler = slog.NewTextHandler(os.Stderr, options)
	}
}

var Logger = slog.New(handler)
