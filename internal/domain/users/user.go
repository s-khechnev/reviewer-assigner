package users

import teamsDomain "reviewer-assigner/internal/domain/teams"

type User struct {
	teamsDomain.Member
	TeamName string
}
