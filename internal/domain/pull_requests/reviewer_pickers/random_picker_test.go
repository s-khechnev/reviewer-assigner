package reviewer_pickers

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	teamsDomain "reviewer-assigner/internal/domain/teams"
	"testing"
)

func TestRandomReviewerPicker_Pick_TableDriven(t *testing.T) {
	picker := &RandomReviewerPicker{}

	testCases := []struct {
		name           string
		members        []teamsDomain.Member
		count          int
		expectedLength int
		expectNil      bool
		description    string
	}{
		{
			name:           "empty_members",
			members:        []teamsDomain.Member{},
			count:          2,
			expectedLength: 0,
			expectNil:      true,
			description:    "should return nil when members slice is empty",
		},
		{
			name: "zero_count",
			members: []teamsDomain.Member{
				{ID: "1", Name: "User1", IsActive: true},
			},
			count:          0,
			expectedLength: 0,
			expectNil:      true,
			description:    "should return nil when count is zero",
		},
		{
			name: "negative_count",
			members: []teamsDomain.Member{
				{ID: "1", Name: "User1", IsActive: true},
			},
			count:          -1,
			expectedLength: 0,
			expectNil:      true,
			description:    "should return nil when count is negative",
		},

		// Equal count cases
		{
			name: "count_equals_members_length",
			members: []teamsDomain.Member{
				{ID: "1", Name: "User1", IsActive: true},
				{ID: "2", Name: "User2", IsActive: true},
			},
			count:          2,
			expectedLength: 2,
			expectNil:      false,
			description:    "should return all members when count equals members length",
		},
		{
			name: "count_greater_than_members_length",
			members: []teamsDomain.Member{
				{ID: "1", Name: "User1", IsActive: true},
				{ID: "2", Name: "User2", IsActive: true},
			},
			count:          5,
			expectedLength: 2,
			expectNil:      false,
			description:    "should return all members when count greater than members length",
		},

		// Normal cases
		{
			name: "count_less_than_members_length",
			members: []teamsDomain.Member{
				{ID: "1", Name: "User1", IsActive: true},
				{ID: "2", Name: "User2", IsActive: true},
				{ID: "3", Name: "User3", IsActive: true},
				{ID: "4", Name: "User4", IsActive: true},
			},
			count:          2,
			expectedLength: 2,
			expectNil:      false,
			description:    "should return exactly count members when count less than members length",
		},
		{
			name: "single_member_exact_count",
			members: []teamsDomain.Member{
				{ID: "1", Name: "User1", IsActive: true},
			},
			count:          1,
			expectedLength: 1,
			expectNil:      false,
			description:    "should return single member when count equals 1",
		},
		{
			name: "single_member_overflow_count",
			members: []teamsDomain.Member{
				{ID: "1", Name: "User1", IsActive: true},
			},
			count:          3,
			expectedLength: 1,
			expectNil:      false,
			description:    "should return single member when count greater than members length",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := picker.Pick(tc.members, tc.count)

			if tc.expectNil {
				assert.Nil(t, result, tc.description)
			} else {
				require.NotNil(t, result, tc.description)
				assert.Len(t, result, tc.expectedLength, tc.description)

				for _, reviewer := range result {
					assert.Contains(t, tc.members, reviewer, "returned member should be from original list")
				}
			}
		})
	}
}
