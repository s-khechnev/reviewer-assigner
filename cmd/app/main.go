package main

import (
	"avito/internal/config"
	"log/slog"
	"os"
)

const (
	envDebug = "debug"
	envProd  = "prod"
)

func main() {
	cfg := config.GetConfig()
	log := initLogger(cfg.Env)

	log.Info("starting service",
		slog.String("env", cfg.Env),
		slog.Int("port", cfg.HttpServer.Port),
	)
	log.Debug("debug messages are enabled")
}

func initLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envDebug:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
