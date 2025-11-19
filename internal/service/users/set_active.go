package users

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	usersDomain "reviewer-assigner/internal/domain/users"
	"reviewer-assigner/internal/logger"
	"reviewer-assigner/internal/service"
	"reviewer-assigner/internal/storage"
)

func (s *UserService) SetIsActive(ctx context.Context, userID string, isActive bool) (*usersDomain.User, error) {
	const op = "services.users.SetIsActive"
	log := s.log.With(
		slog.String("op", op),
		slog.String("user_id", userID),
	)

	user, err := s.userRepo.SetIsActive(ctx, userID, isActive)
	if errors.Is(err, storage.ErrUserNotFound) {
		log.Warn("user not found")

		return nil, service.ErrUserNotFound
	}
	if err != nil {
		log.Error("failed to set is active", logger.ErrAttr(err))

		return nil, fmt.Errorf("failed to set is active: %w", service.ErrUserNotFound)
	}

	log.Info("success set is active", slog.Any("user", *user))

	return user, nil
}
