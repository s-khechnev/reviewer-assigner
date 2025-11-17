package users

import (
	"context"
	"fmt"
	usersDomain "reviewer-assigner/internal/domain/users"
	"reviewer-assigner/internal/service"
)

func (s *UserService) SetIsActive(ctx context.Context, userID string, isActive bool) (*usersDomain.User, error) {
	user, err := s.userRepo.SetIsActive(ctx, userID, isActive)
	if err != nil {
		return nil, fmt.Errorf("failed to set is active: %w", service.ErrUserNotFound)
	}

	return user, nil
}
