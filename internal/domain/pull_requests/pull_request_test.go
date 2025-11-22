package pull_requests

import (
	"errors"
	"reviewer-assigner/internal/domain"
	"testing"
	"time"

	teamsDomain "reviewer-assigner/internal/domain/teams"
)

// Mock implementations for ReviewerPicker and ReviewerReassigner
type MockReviewerPicker struct {
	PickFunc func(members []teamsDomain.Member, count int) []teamsDomain.Member
}

func (m *MockReviewerPicker) Pick(members []teamsDomain.Member, count int) []teamsDomain.Member {
	if m.PickFunc != nil {
		return m.PickFunc(members, count)
	}
	return nil
}

type MockReviewerReassigner struct {
	ReassignFunc func(oldReviewer *teamsDomain.Member, members []teamsDomain.Member) (newReviewer *teamsDomain.Member, err error)
}

func (m *MockReviewerReassigner) Reassign(oldReviewer *teamsDomain.Member, members []teamsDomain.Member) (*teamsDomain.Member, error) {
	if m.ReassignFunc != nil {
		return m.ReassignFunc(oldReviewer, members)
	}
	return nil, nil
}

func TestPullRequest_AssignReviewers(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name              string
		pr                *PullRequest
		members           []teamsDomain.Member
		picker            ReviewerPicker
		count             int
		wantErr           bool
		expectedReviewers []string
		expectedError     error
	}{
		{
			name: "success: assign two active reviewers excluding author",
			pr: &PullRequest{
				PullRequestShort: PullRequestShort{
					AuthorID: "author1",
					Status:   StatusOpen,
				},
			},
			members: []teamsDomain.Member{
				{ID: "author1", Name: "Author", IsActive: true},
				{ID: "reviewer1", Name: "Reviewer1", IsActive: true},
				{ID: "reviewer2", Name: "Reviewer2", IsActive: true},
				{ID: "inactive1", Name: "Inactive1", IsActive: false},
			},
			picker: &MockReviewerPicker{
				PickFunc: func(members []teamsDomain.Member, count int) []teamsDomain.Member {
					return members[:2] // Return first two active members
				},
			},
			count:             2,
			wantErr:           false,
			expectedReviewers: []string{"reviewer1", "reviewer2"},
		},
		{
			name: "success: only one active reviewer available",
			pr: &PullRequest{
				PullRequestShort: PullRequestShort{
					AuthorID: "author1",
					Status:   StatusOpen,
				},
			},
			members: []teamsDomain.Member{
				{ID: "author1", Name: "Author", IsActive: true},
				{ID: "reviewer1", Name: "Reviewer1", IsActive: true},
				{ID: "inactive1", Name: "Inactive1", IsActive: false},
			},
			picker: &MockReviewerPicker{
				PickFunc: func(members []teamsDomain.Member, count int) []teamsDomain.Member {
					return members[:1] // Return only one available
				},
			},
			count:             2,
			wantErr:           false,
			expectedReviewers: []string{"reviewer1"},
		},
		{
			name: "success: no active reviewers available",
			pr: &PullRequest{
				PullRequestShort: PullRequestShort{
					AuthorID: "author1",
					Status:   StatusOpen,
				},
			},
			members: []teamsDomain.Member{
				{ID: "author1", Name: "Author", IsActive: true},
				{ID: "inactive1", Name: "Inactive1", IsActive: false},
			},
			picker: &MockReviewerPicker{
				PickFunc: func(members []teamsDomain.Member, count int) []teamsDomain.Member {
					return []teamsDomain.Member{} // No available reviewers
				},
			},
			count:             2,
			wantErr:           false,
			expectedReviewers: []string{},
		},
		{
			name: "error: cannot assign reviewers to merged PR",
			pr: &PullRequest{
				PullRequestShort: PullRequestShort{
					AuthorID: "author1",
					Status:   StatusMerged,
				},
				MergedAt: &now,
			},
			members: []teamsDomain.Member{
				{ID: "author1", Name: "Author", IsActive: true},
				{ID: "reviewer1", Name: "Reviewer1", IsActive: true},
			},
			picker: &MockReviewerPicker{
				PickFunc: func(members []teamsDomain.Member, count int) []teamsDomain.Member {
					return members[:1]
				},
			},
			count:         2,
			wantErr:       true,
			expectedError: domain.ErrPullRequestAlreadyMerged,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.pr.AssignReviewers(tt.members, tt.picker, tt.count)

			if (err != nil) != tt.wantErr {
				t.Errorf("AssignReviewers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && !errors.Is(err, tt.expectedError) {
				t.Errorf("AssignReviewers() error = %v, expectedError %v", err, tt.expectedError)
				return
			}

			if !tt.wantErr {
				if len(tt.pr.AssignedReviewers) != len(tt.expectedReviewers) {
					t.Errorf("AssignReviewers() reviewers count = %v, want %v",
						len(tt.pr.AssignedReviewers), len(tt.expectedReviewers))
				}

				for i, reviewer := range tt.expectedReviewers {
					if tt.pr.AssignedReviewers[i] != reviewer {
						t.Errorf("AssignReviewers() reviewer[%d] = %v, want %v",
							i, tt.pr.AssignedReviewers[i], reviewer)
					}
				}
			}
		})
	}
}

func TestPullRequest_Merge(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name        string
		pr          *PullRequest
		wantErr     bool
		expectedErr error
	}{
		{
			name: "success: merge open PR",
			pr: &PullRequest{
				PullRequestShort: PullRequestShort{
					Status: StatusOpen,
				},
			},
			wantErr: false,
		},
		{
			name: "error: already merged PR",
			pr: &PullRequest{
				PullRequestShort: PullRequestShort{
					Status: StatusMerged,
				},
				MergedAt: &now,
			},
			wantErr:     true,
			expectedErr: domain.ErrPullRequestAlreadyMerged,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.pr.Merge()

			if (err != nil) != tt.wantErr {
				t.Errorf("Merge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && !errors.Is(err, tt.expectedErr) {
				t.Errorf("Merge() error = %v, expectedError %v", err, tt.expectedErr)
				return
			}

			if !tt.wantErr {
				if tt.pr.Status != StatusMerged {
					t.Errorf("Merge() status = %v, want %v", tt.pr.Status, StatusMerged)
				}
				if tt.pr.MergedAt == nil {
					t.Error("Merge() MergedAt should not be nil")
				}
			}
		})
	}
}

func TestPullRequest_Reassign(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name              string
		pr                *PullRequest
		oldReviewer       *teamsDomain.Member
		members           []teamsDomain.Member
		reassigner        ReviewerReassigner
		wantErr           bool
		expectedError     error
		expectedReviewers []string
	}{
		{
			name: "success: reassign reviewer",
			pr: &PullRequest{
				PullRequestShort: PullRequestShort{
					AuthorID: "author1",
					Status:   StatusOpen,
				},
				AssignedReviewers: []string{"reviewer1", "reviewer2"},
			},
			oldReviewer: &teamsDomain.Member{ID: "reviewer1", Name: "Reviewer1", IsActive: true},
			members: []teamsDomain.Member{
				{ID: "author1", Name: "Author", IsActive: true},
				{ID: "reviewer1", Name: "Reviewer1", IsActive: true},
				{ID: "reviewer2", Name: "Reviewer2", IsActive: true},
				{ID: "reviewer3", Name: "Reviewer3", IsActive: true},
				{ID: "inactive1", Name: "Inactive1", IsActive: false},
			},
			reassigner: &MockReviewerReassigner{
				ReassignFunc: func(oldReviewer *teamsDomain.Member, members []teamsDomain.Member) (*teamsDomain.Member, error) {
					// Should receive only active members excluding author, old reviewer, and already assigned
					if len(members) != 1 || members[0].ID != "reviewer3" {
						t.Errorf("Reassign received wrong members list: %v", members)
					}
					return &teamsDomain.Member{ID: "reviewer3", Name: "Reviewer3", IsActive: true}, nil
				},
			},
			wantErr:           false,
			expectedReviewers: []string{"reviewer3", "reviewer2"},
		},
		{
			name: "error: no available candidates for reassignment",
			pr: &PullRequest{
				PullRequestShort: PullRequestShort{
					AuthorID: "author1",
					Status:   StatusOpen,
				},
				AssignedReviewers: []string{"reviewer1", "reviewer2"},
			},
			oldReviewer: &teamsDomain.Member{ID: "reviewer1", Name: "Reviewer1", IsActive: true},
			members: []teamsDomain.Member{
				{ID: "author1", Name: "Author", IsActive: true},
				{ID: "reviewer1", Name: "Reviewer1", IsActive: true},
				{ID: "reviewer2", Name: "Reviewer2", IsActive: true},
				{ID: "inactive1", Name: "Inactive1", IsActive: false},
			},
			reassigner: &MockReviewerReassigner{
				ReassignFunc: func(oldReviewer *teamsDomain.Member, members []teamsDomain.Member) (*teamsDomain.Member, error) {
					return nil, errors.New("no available candidates")
				},
			},
			wantErr: true,
		},
		{
			name: "error: cannot reassign on merged PR",
			pr: &PullRequest{
				PullRequestShort: PullRequestShort{
					AuthorID: "author1",
					Status:   StatusMerged,
				},
				AssignedReviewers: []string{"reviewer1", "reviewer2"},
				MergedAt:          &now,
			},
			oldReviewer: &teamsDomain.Member{ID: "reviewer1", Name: "Reviewer1", IsActive: true},
			members: []teamsDomain.Member{
				{ID: "reviewer3", Name: "Reviewer3", IsActive: true},
			},
			reassigner: &MockReviewerReassigner{
				ReassignFunc: func(oldReviewer *teamsDomain.Member, members []teamsDomain.Member) (*teamsDomain.Member, error) {
					return &teamsDomain.Member{ID: "reviewer3", Name: "Reviewer3", IsActive: true}, nil
				},
			},
			wantErr:       true,
			expectedError: domain.ErrPullRequestAlreadyMerged,
		},
		{
			name: "success: reassign when old reviewer not found in assigned reviewers",
			pr: &PullRequest{
				PullRequestShort: PullRequestShort{
					AuthorID: "author1",
					Status:   StatusOpen,
				},
				AssignedReviewers: []string{"reviewer1", "reviewer2"},
			},
			oldReviewer: &teamsDomain.Member{ID: "reviewer3", Name: "Reviewer3", IsActive: true},
			members: []teamsDomain.Member{
				{ID: "reviewer4", Name: "Reviewer4", IsActive: true},
			},
			reassigner: &MockReviewerReassigner{
				ReassignFunc: func(oldReviewer *teamsDomain.Member, members []teamsDomain.Member) (*teamsDomain.Member, error) {
					return &teamsDomain.Member{ID: "reviewer4", Name: "Reviewer4", IsActive: true}, nil
				},
			},
			wantErr:           false,
			expectedReviewers: []string{"reviewer1", "reviewer2"}, // Should remain unchanged
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.pr.Reassign(tt.oldReviewer, tt.members, tt.reassigner)

			if (err != nil) != tt.wantErr {
				t.Errorf("Reassign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.expectedError != nil && !errors.Is(err, tt.expectedError) {
				t.Errorf("Reassign() error = %v, expectedError %v", err, tt.expectedError)
				return
			}

			if !tt.wantErr {
				if len(tt.pr.AssignedReviewers) != len(tt.expectedReviewers) {
					t.Errorf("Reassign() reviewers count = %v, want %v",
						len(tt.pr.AssignedReviewers), len(tt.expectedReviewers))
					return
				}

				for i, reviewer := range tt.expectedReviewers {
					if tt.pr.AssignedReviewers[i] != reviewer {
						t.Errorf("Reassign() reviewer[%d] = %v, want %v",
							i, tt.pr.AssignedReviewers[i], reviewer)
					}
				}
			}
		})
	}
}
