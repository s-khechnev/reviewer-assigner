package pull_requests

import (
	"reviewer-assigner/internal/domain"
	teamsDomain "reviewer-assigner/internal/domain/teams"
	"slices"
	"time"
)

type StatusPR string

const (
	StatusOpen   StatusPR = "OPEN"
	StatusMerged StatusPR = "MERGED"
)

type PullRequestShort struct {
	ID       string
	Name     string
	AuthorID string
	Status   StatusPR
}

type PullRequest struct {
	PullRequestShort
	AssignedReviewers []string
	CreatedAt         *time.Time
	MergedAt          *time.Time
}

type ReviewerPicker interface {
	Pick(members []teamsDomain.Member, count int) []teamsDomain.Member
}

type ReviewerReassigner interface {
	Reassign(oldReviewer *teamsDomain.Member, members []teamsDomain.Member) (newReviewer *teamsDomain.Member, err error)
}

func (p *PullRequest) AssignReviewers(members []teamsDomain.Member, picker ReviewerPicker, count int) error {
	if p.Status == StatusMerged {
		return domain.ErrPullRequestAlreadyMerged
	}

	const activeMembersDefaultCap = 2
	activeMembersExcludeAuthor := make([]teamsDomain.Member, 0, activeMembersDefaultCap)
	for _, member := range members {
		if member.IsActive && member.ID != p.AuthorID {
			activeMembersExcludeAuthor = append(activeMembersExcludeAuthor, member)
		}
	}

	reviewers := picker.Pick(activeMembersExcludeAuthor, count)

	reviewerIDs := make([]string, 0, len(reviewers))
	for _, reviewer := range reviewers {
		reviewerIDs = append(reviewerIDs, reviewer.ID)
	}

	p.AssignedReviewers = reviewerIDs

	return nil
}

func (p *PullRequest) Merge() error {
	if p.Status == StatusMerged {
		return domain.ErrPullRequestAlreadyMerged
	}

	p.Status = StatusMerged
	now := time.Now()
	p.MergedAt = &now

	return nil
}

func (p *PullRequest) Reassign(
	oldReviewer *teamsDomain.Member,
	members []teamsDomain.Member,
	reassigner ReviewerReassigner,
) (string, error) {
	if p.Status == StatusMerged {
		return "", domain.ErrPullRequestAlreadyMerged
	}

	isAlreadyReviewer := func(member *teamsDomain.Member) bool {
		return slices.Index(p.AssignedReviewers, member.ID) != -1
	}

	const activeMembersDefaultCap = 2
	activeMembersExcludeAuthorReviewers := make([]teamsDomain.Member, 0, activeMembersDefaultCap)
	for _, member := range members {
		if member.IsActive &&
			member.ID != oldReviewer.ID &&
			member.ID != p.AuthorID &&
			!isAlreadyReviewer(&member) {
			activeMembersExcludeAuthorReviewers = append(activeMembersExcludeAuthorReviewers, member)
		}
	}

	newReviewer, err := reassigner.Reassign(oldReviewer, activeMembersExcludeAuthorReviewers)
	if err != nil {
		return "", err
	}

	for i, reviewerID := range p.AssignedReviewers {
		if reviewerID == oldReviewer.ID {
			p.AssignedReviewers[i] = newReviewer.ID
			break
		}
	}

	return newReviewer.ID, nil
}
