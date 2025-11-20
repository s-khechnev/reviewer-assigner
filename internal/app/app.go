package app

import (
	"context"
	"errors"
	"fmt"
	trmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/gin-gonic/gin"
	sloggin "github.com/samber/slog-gin"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"reviewer-assigner/internal/config"
	reviewerPicker "reviewer-assigner/internal/domain/pull_requests/reviewer_pickers"
	reviewerAssigner "reviewer-assigner/internal/domain/pull_requests/reviewer_reassigners"
	prHandler "reviewer-assigner/internal/http/handlers/pull_requests"
	teamsHandler "reviewer-assigner/internal/http/handlers/teams"
	usersHandler "reviewer-assigner/internal/http/handlers/users"
	"reviewer-assigner/internal/logger"
	prService "reviewer-assigner/internal/service/pull_requests"
	teamsService "reviewer-assigner/internal/service/teams"
	usersService "reviewer-assigner/internal/service/users"
	"reviewer-assigner/internal/storage/postgres"
	pullRequestsRepo "reviewer-assigner/internal/storage/pull_requests"
	teamsRepo "reviewer-assigner/internal/storage/teams"
	usersRepo "reviewer-assigner/internal/storage/users"
	"syscall"
	"time"

	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
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

	txManager := manager.Must(trmpgx.NewDefaultFactory(pool))

	teamRepo := teamsRepo.NewPostgresTeamRepository(pool, trmpgx.DefaultCtxGetter)
	userRepo := usersRepo.NewPostgresUserRepository(pool, trmpgx.DefaultCtxGetter)
	pullRequestRepo := pullRequestsRepo.NewPostgresPullRequestRepository(pool, trmpgx.DefaultCtxGetter)

	teamService := teamsService.NewTeamService(log, teamRepo, txManager)
	userService := usersService.NewUserService(log, userRepo, pullRequestRepo, txManager)
	pullRequestService := prService.NewPullRequestService(
		log,
		userRepo,
		teamRepo,
		pullRequestRepo,
		&reviewerPicker.RandomReviewerPicker{},
		reviewerAssigner.NewRandomReviewerReassigner(),
		txManager,
	)

	teamHandler := teamsHandler.NewTeamHandler(log, teamService)
	userHandler := usersHandler.NewUserHandler(log, userService)
	pullRequestHandler := prHandler.NewPullRequestHandler(log, pullRequestService)

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

	{
		userGroup := r.Group("/users")
		userGroup.POST("/setIsActive", userHandler.SetIsActive)
		userGroup.GET("/getReview", userHandler.GetReview)
	}

	{
		pullRequestGroup := r.Group("/pullRequest")
		pullRequestGroup.POST("/create", pullRequestHandler.Create)
		pullRequestGroup.POST("/merge", pullRequestHandler.Merge)
		pullRequestGroup.POST("/reassign", pullRequestHandler.Reassign)
	}

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.HttpServer.Address, cfg.HttpServer.Port),
		Handler:      r,
		IdleTimeout:  cfg.HttpServer.IdleTimeout,
		WriteTimeout: cfg.HttpServer.Timeout,
		ReadTimeout:  cfg.HttpServer.Timeout,
	}

	go func() {
		if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("failed listen", logger.ErrAttr(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down service")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = server.Shutdown(ctx); err != nil {
		log.Error("failed to shutdown", logger.ErrAttr(err))
	}
}
