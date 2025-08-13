package slogger

import (
	"log/slog"
	"os"
)

func SetupLogger() *slog.Logger {
	log := setupPrettySlog()
	return log
}

func setupPrettySlog() *slog.Logger {
	opts := PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
