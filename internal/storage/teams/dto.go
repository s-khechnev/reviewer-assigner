package teams

import (
	teamsDomain "reviewer-assigner/internal/domain/teams"
)

type MemberDB struct {
	ID       string `db:"id"`
	MemberID string `db:"user_id"`
	Name     string `db:"name"`
	IsActive bool   `db:"is_active"`
}

func DbToDomainMember(d *MemberDB) *teamsDomain.Member {
	return &teamsDomain.Member{
		ID:       d.MemberID,
		Name:     d.Name,
		IsActive: d.IsActive,
	}
}
