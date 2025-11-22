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

func (m *Member) Equal(o *Member) bool {
	if m.IsActive != o.IsActive {
		return false
	}
	if m.ID != o.ID {
		return false
	}
	if m.Name != o.Name {
		return false
	}

	return true
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
			// check if really updated
			if !t.Members[idx].Equal(&updated) {
				t.Members[idx] = updated
			}
		}
	}

	return nil
}

func hasSameMemberIDs(oldMembers, newMembers []Member) bool {
	if len(newMembers) == 0 && len(oldMembers) != 0 {
		return false
	}

	for _, newMember := range newMembers {
		if !slices.ContainsFunc(oldMembers, func(m Member) bool {
			return m.ID == newMember.ID
		}) {
			return false
		}
	}

	return true
}
