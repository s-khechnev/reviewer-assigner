package teams

import (
	"reviewer-assigner/internal/domain"
	"slices"
)

type Member struct {
	ID       string
	Name     string
	IsActive bool
}

type Team struct {
	Name    string
	Members []Member
}

func (t *Team) UpdateMembers(updatedMembers []Member) error {
	if !hasSameMemberIDs(t.Members, updatedMembers) {
		return domain.ErrTeamMembersMismatch
	}

	for _, updated := range updatedMembers {
		idx := slices.IndexFunc(t.Members, func(m Member) bool {
			return m.ID == updated.ID
		})
		if idx != -1 {
			t.Members[idx] = updated
		}
	}

	return nil
}

func hasSameMemberIDs(oldMembers, newMembers []Member) bool {
	for _, newMember := range newMembers {
		if !slices.ContainsFunc(oldMembers, func(m Member) bool {
			return m.ID == newMember.ID
		}) {
			return false
		}
	}

	return true
}
