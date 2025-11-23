package main

import (
	"context"
	"log/slog"
	"reviewer-assigner/internal/app"
	"reviewer-assigner/internal/config"
	"reviewer-assigner/internal/logger"
)

func main() {
	ctx := context.Background()

	cfg := config.Must()
	log := logger.New(cfg.Env)

	log.Info("starting service",
		slog.String("env", cfg.Env),
		slog.Int("port", cfg.HTTPServer.Port),
	)
	log.Debug("debug messages are enabled")

	app.Run(ctx, cfg, log)

	log.Info("service exiting")
}
