package users

import teamsDomain "reviewer-assigner/internal/domain/teams"

type User struct {
	teamsDomain.Member

	TeamName string
}

func (u *User) SetIsActive(isActive bool) error {
	// if user is reviewer should we reassign???
	u.IsActive = isActive

	return nil
}
