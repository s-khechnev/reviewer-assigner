package teams

import (
	"errors"
	"reviewer-assigner/internal/domain"
	"slices"
	"testing"
)

func TestTeam_UpdateMembers(t *testing.T) {
	// Создаем тестовые данные
	baseMembers := []Member{
		{ID: "1", Name: "Alice", IsActive: true},
		{ID: "2", Name: "Bob", IsActive: true},
		{ID: "3", Name: "Charlie", IsActive: false},
	}

	tests := []struct {
		name           string
		initialMembers []Member
		updatedMembers []Member
		wantErr        error
		wantMembers    []Member
	}{
		{
			name:           "successful update - update names and status",
			initialMembers: baseMembers,
			updatedMembers: []Member{
				{ID: "1", Name: "Alice Smith", IsActive: true},
				{ID: "2", Name: "Bob Johnson", IsActive: false},
				{ID: "3", Name: "Charlie Brown", IsActive: true},
			},
			wantErr: nil,
			wantMembers: []Member{
				{ID: "1", Name: "Alice Smith", IsActive: true},
				{ID: "2", Name: "Bob Johnson", IsActive: false},
				{ID: "3", Name: "Charlie Brown", IsActive: true},
			},
		},
		{
			name:           "successful update - partial updates",
			initialMembers: baseMembers,
			updatedMembers: []Member{
				{ID: "1", Name: "Alice", IsActive: false},      // только статус изменился
				{ID: "2", Name: "Bob", IsActive: true},         // без изменений
				{ID: "3", Name: "Charlie New", IsActive: true}, // имя и статус
			},
			wantErr: nil,
			wantMembers: []Member{
				{ID: "1", Name: "Alice", IsActive: false},
				{ID: "2", Name: "Bob", IsActive: true},
				{ID: "3", Name: "Charlie New", IsActive: true},
			},
		},
		{
			name:           "error - missing member in updated list",
			initialMembers: baseMembers,
			updatedMembers: []Member{
				{ID: "1", Name: "Alice", IsActive: true},
				{ID: "2", Name: "Bob", IsActive: true},
				// ID "3" отсутствует
			},
			wantErr:     nil,
			wantMembers: baseMembers, // члены не должны измениться
		},
		{
			name:           "error - extra member in updated list",
			initialMembers: baseMembers,
			updatedMembers: []Member{
				{ID: "1", Name: "Alice", IsActive: true},
				{ID: "2", Name: "Bob", IsActive: true},
				{ID: "3", Name: "Charlie", IsActive: true},
				{ID: "4", Name: "David", IsActive: true}, // лишний член
			},
			wantErr:     domain.ErrTeamMembersMismatch,
			wantMembers: baseMembers, // члены не должны измениться
		},
		{
			name:           "error - completely different members",
			initialMembers: baseMembers,
			updatedMembers: []Member{
				{ID: "4", Name: "David", IsActive: true},
				{ID: "5", Name: "Eve", IsActive: true},
			},
			wantErr:     domain.ErrTeamMembersMismatch,
			wantMembers: baseMembers,
		},
		{
			name:           "empty team - update with empty list",
			initialMembers: []Member{},
			updatedMembers: []Member{},
			wantErr:        nil,
			wantMembers:    []Member{},
		},
		{
			name:           "empty team - update with non-empty list",
			initialMembers: []Member{},
			updatedMembers: []Member{
				{ID: "1", Name: "Alice", IsActive: true},
			},
			wantErr:     domain.ErrTeamMembersMismatch,
			wantMembers: []Member{},
		},
		{
			name:           "single member team - successful update",
			initialMembers: []Member{{ID: "1", Name: "Alice", IsActive: true}},
			updatedMembers: []Member{{ID: "1", Name: "Alice Smith", IsActive: false}},
			wantErr:        nil,
			wantMembers:    []Member{{ID: "1", Name: "Alice Smith", IsActive: false}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			team := &Team{
				Name:    "Test Team",
				Members: slices.Clone(tt.initialMembers),
			}

			err := team.UpdateMembers(tt.updatedMembers)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("UpdateMembers() error = %v, wantErr %v", err, tt.wantErr)
			}

			if len(team.Members) != len(tt.wantMembers) {
				t.Errorf(
					"UpdateMembers() member count = %d, want %d",
					len(team.Members),
					len(tt.wantMembers),
				)
			}

			for i, wantMember := range tt.wantMembers {
				if i >= len(team.Members) {
					break
				}
				gotMember := team.Members[i]
				if gotMember.ID != wantMember.ID || gotMember.Name != wantMember.Name ||
					gotMember.IsActive != wantMember.IsActive {
					t.Errorf("UpdateMembers() member[%d] = %+v, want %+v", i, gotMember, wantMember)
				}
			}
		})
	}
}

func Test_hasSameMemberIDs(t *testing.T) {
	tests := []struct {
		name        string
		oldMembers  []Member
		newMembers  []Member
		expected    bool
		description string
	}{
		{
			name: "same_members_same_order",
			oldMembers: []Member{
				{ID: "1", Name: "A", IsActive: true},
				{ID: "2", Name: "B", IsActive: true},
			},
			newMembers: []Member{
				{ID: "1", Name: "A Updated", IsActive: false},
				{ID: "2", Name: "B Updated", IsActive: false},
			},
			expected:    true,
			description: "Одинаковые ID в том же порядке",
		},
		{
			name: "same_members_different_order",
			oldMembers: []Member{
				{ID: "1", Name: "A", IsActive: true},
				{ID: "2", Name: "B", IsActive: true},
			},
			newMembers: []Member{
				{ID: "2", Name: "B Updated", IsActive: false},
				{ID: "1", Name: "A Updated", IsActive: false},
			},
			expected:    true,
			description: "Одинаковые ID в разном порядке",
		},
		{
			name: "different_members",
			oldMembers: []Member{
				{ID: "1", Name: "A", IsActive: true},
				{ID: "2", Name: "B", IsActive: true},
			},
			newMembers: []Member{
				{ID: "1", Name: "A", IsActive: true},
				{ID: "3", Name: "C", IsActive: true},
			},
			expected:    false,
			description: "Разные наборы ID",
		},
		{
			name:        "both_empty",
			oldMembers:  []Member{},
			newMembers:  []Member{},
			expected:    true,
			description: "Оба списка пусты",
		},
		{
			name:       "old_empty_new_not",
			oldMembers: []Member{},
			newMembers: []Member{
				{ID: "1", Name: "A", IsActive: true},
			},
			expected:    false,
			description: "Старый список пуст, новый - нет",
		},
		{
			name: "old_not_empty_new_empty",
			oldMembers: []Member{
				{ID: "1", Name: "A", IsActive: true},
			},
			newMembers:  []Member{},
			expected:    false,
			description: "Новый список пуст, старый - нет",
		},
		{
			name: "new_has_extra_member",
			oldMembers: []Member{
				{ID: "1", Name: "A", IsActive: true},
			},
			newMembers: []Member{
				{ID: "1", Name: "A", IsActive: true},
				{ID: "2", Name: "B", IsActive: true},
			},
			expected:    false,
			description: "Новый список содержит дополнительного участника",
		},
		{
			name: "new_missing_member",
			oldMembers: []Member{
				{ID: "1", Name: "A", IsActive: true},
				{ID: "2", Name: "B", IsActive: true},
			},
			newMembers: []Member{
				{ID: "1", Name: "A", IsActive: true},
			},
			expected:    true,
			description: "Новый список не содержит всех участников",
		},
		{
			name: "case_sensitive_ids",
			oldMembers: []Member{
				{ID: "ABC", Name: "A", IsActive: true},
			},
			newMembers: []Member{
				{ID: "abc", Name: "A", IsActive: true},
			},
			expected:    false,
			description: "ID чувствительны к регистру",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasSameMemberIDs(tt.oldMembers, tt.newMembers)
			if result != tt.expected {
				t.Errorf("hasSameMemberIDs() = %v, expected %v", result, tt.expected)
				t.Errorf("Старые участники: %v", tt.oldMembers)
				t.Errorf("Новые участники: %v", tt.newMembers)
			}
		})
	}
}
