package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	sloggin "github.com/samber/slog-gin"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"reviewer-assigner/internal/config"
	teamsHandler "reviewer-assigner/internal/http/handlers/teams"
	"reviewer-assigner/internal/logger"
	teamsService "reviewer-assigner/internal/service/teams"
	"reviewer-assigner/internal/storage/postgres"
	teamsRepo "reviewer-assigner/internal/storage/teams"
	"syscall"
	"time"
)

func Run(ctx context.Context, cfg *config.Config, log *slog.Logger) {
	dsnConnString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name, cfg.DB.SslMode)

	pool, err := postgres.NewPool(ctx, dsnConnString)
	if err != nil {
		log.Error("failed to create pool to database", logger.ErrAttr(err))
		return
	}

	teamRepo := teamsRepo.NewPostgresTeamRepository(pool)
	teamService := teamsService.NewTeamService(log, teamRepo)
	teamHandler := teamsHandler.NewTeamHandler(log, teamService)

	r := gin.New()

	switch cfg.Env {
	case config.EnvProd:
		gin.SetMode(gin.ReleaseMode)
	case config.EnvDebug:
		gin.SetMode(gin.DebugMode)
	}

	r.Use(gin.Recovery())
	r.Use(sloggin.New(log))

	{
		teamGroup := r.Group("/team")
		teamGroup.POST("/add", teamHandler.AddTeam)
		teamGroup.GET("/get", teamHandler.GetTeam)
	}

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.HttpServer.Address, cfg.HttpServer.Port),
		Handler:      r,
		IdleTimeout:  cfg.HttpServer.IdleTimeout,
		WriteTimeout: cfg.HttpServer.Timeout,
		ReadTimeout:  cfg.HttpServer.Timeout,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("failed listen", logger.ErrAttr(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down service")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("failed to shutdown", logger.ErrAttr(err))
	}
}
