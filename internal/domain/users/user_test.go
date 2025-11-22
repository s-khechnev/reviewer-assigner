package users

import (
	teamsDomain "reviewer-assigner/internal/domain/teams"
	"testing"
)

func TestUser_SetIsActive(t *testing.T) {
	tests := []struct {
		name           string
		initialUser    *User
		isActive       bool
		expectedActive bool
	}{
		{
			name: "activate inactive user",
			initialUser: &User{
				Member: teamsDomain.Member{
					ID:       "1",
					Name:     "Alice",
					IsActive: false,
				},
				TeamName: "Team A",
			},
			isActive:       true,
			expectedActive: true,
		},
		{
			name: "deactivate active user",
			initialUser: &User{
				Member: teamsDomain.Member{
					ID:       "2",
					Name:     "Bob",
					IsActive: true,
				},
				TeamName: "Team B",
			},
			isActive:       false,
			expectedActive: false,
		},
		{
			name: "set active to same value - true",
			initialUser: &User{
				Member: teamsDomain.Member{
					ID:       "3",
					Name:     "Charlie",
					IsActive: true,
				},
				TeamName: "Team C",
			},
			isActive:       true,
			expectedActive: true,
		},
		{
			name: "set active to same value - false",
			initialUser: &User{
				Member: teamsDomain.Member{
					ID:       "4",
					Name:     "David",
					IsActive: false,
				},
				TeamName: "Team D",
			},
			isActive:       false,
			expectedActive: false,
		},
		{
			name: "user with empty fields",
			initialUser: &User{
				Member: teamsDomain.Member{
					ID:       "",
					Name:     "",
					IsActive: true,
				},
				TeamName: "",
			},
			isActive:       false,
			expectedActive: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{
				Member:   tt.initialUser.Member,
				TeamName: tt.initialUser.TeamName,
			}

			_ = user.SetIsActive(tt.isActive)

			if user.IsActive != tt.expectedActive {
				t.Errorf(
					"SetIsActive() IsActive = %v, expected %v",
					user.IsActive,
					tt.expectedActive,
				)
			}

			if user.ID != tt.initialUser.ID {
				t.Errorf("SetIsActive() changed ID from %s to %s", tt.initialUser.ID, user.ID)
			}
			if user.Name != tt.initialUser.Name {
				t.Errorf("SetIsActive() changed Name from %s to %s", tt.initialUser.Name, user.Name)
			}
			if user.TeamName != tt.initialUser.TeamName {
				t.Errorf(
					"SetIsActive() changed TeamName from %s to %s",
					tt.initialUser.TeamName,
					user.TeamName,
				)
			}
		})
	}
}
