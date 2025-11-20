package users

import (
	teamsDomain "reviewer-assigner/internal/domain/teams"
	usersDomain "reviewer-assigner/internal/domain/users"
)

type UserDB struct {
	ID       int64  `db:"id"`
	UserID   string `db:"user_id"`
	Name     string `db:"name"`
	IsActive bool   `db:"is_active"`
	TeamName string `db:"team_name"`
}

func toDomainUser(u *UserDB) *usersDomain.User {
	return &usersDomain.User{
		Member: teamsDomain.Member{
			ID:       u.UserID,
			Name:     u.Name,
			IsActive: u.IsActive,
		},
		TeamName: u.TeamName,
	}
}
