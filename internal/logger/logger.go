package logger

import (
	"log/slog"
	"os"
	"reviewer-assigner/internal/config"
)

func New(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case config.EnvDebug:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case config.EnvProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}),
		)
	default:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}),
		)
	}

	return log
}

func ErrAttr(err error) slog.Attr {
	return slog.Any("error", err)
}
