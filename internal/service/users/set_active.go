package users

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	usersDomain "reviewer-assigner/internal/domain/users"
	"reviewer-assigner/internal/logger"
	"reviewer-assigner/internal/service"
)

func (s *UserService) SetIsActive(
	ctx context.Context,
	userID string,
	isActive bool,
) (user *usersDomain.User, err error) {
	const op = "services.users.SetIsActive"
	log := s.log.With(
		slog.String("op", op),
		slog.String("user_id", userID),
	)

	err = s.txManager.Do(ctx, func(ctx context.Context) error {
		user, err = s.userRepo.GetUserByID(ctx, userID)
		if errors.Is(err, service.ErrUserNotFound) {
			log.Warn("user not found")

			return service.ErrUserNotFound
		}
		if err != nil {
			log.Error("failed to get user", logger.ErrAttr(err))

			return fmt.Errorf("failed to get user: %w", err)
		}

		log.Info("got user", slog.Any("user", user))

		err = user.SetIsActive(isActive)
		if err != nil {
			log.Error("failed to set is active", logger.ErrAttr(err))

			return fmt.Errorf("failed to set is active: %w", err)
		}

		log.Info("update is active", slog.Bool("is_active", user.IsActive))

		err = s.userRepo.UpdateIsActive(ctx, user)
		if err != nil {
			log.Error("failed to update is active", logger.ErrAttr(err))

			return fmt.Errorf("failed to update is active: %w", service.ErrUserNotFound)
		}

		log.Info("user saved")

		return nil
	})

	return user, err
}
