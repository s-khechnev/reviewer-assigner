package main

import (
	"log/slog"
	"reviewer-assigner/internal/app"
	"reviewer-assigner/internal/config"
	"reviewer-assigner/internal/logger"
)

func main() {
	cfg := config.GetConfig()
	log := logger.New(cfg.Env)

	log.Info("starting service",
		slog.String("env", cfg.Env),
		slog.Int("port", cfg.HttpServer.Port),
	)
	log.Debug("debug messages are enabled")

	app.Run(cfg, log)
}
