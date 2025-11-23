package stats

import "reviewer-assigner/internal/http/handlers/stats"

type UserAssignmentDB struct {
	UserID          string `db:"user_id"`
	Name            string `db:"username"`
	AssignmentCount int    `db:"assignment_count"`
}

func DBToDomainUserAssignment(u *UserAssignmentDB) stats.UserAssignment {
	return stats.UserAssignment{
		UserID: u.UserID,
		Name:   u.Name,
		Count:  u.AssignmentCount,
	}
}
