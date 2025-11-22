package reassigners

import (
	"reviewer-assigner/internal/domain"
	reviewerPickers "reviewer-assigner/internal/domain/pullrequests/pickers"
	teamsDomain "reviewer-assigner/internal/domain/teams"
)

type RandomReviewerReassigner struct {
	picker *reviewerPickers.RandomReviewerPicker
}

func NewRandomReviewerReassigner() *RandomReviewerReassigner {
	return &RandomReviewerReassigner{
		picker: &reviewerPickers.RandomReviewerPicker{},
	}
}

func (r *RandomReviewerReassigner) Reassign(
	_ *teamsDomain.Member,
	members []teamsDomain.Member,
) (*teamsDomain.Member, error) {
	reviewers := r.picker.Pick(members, 1)
	if len(reviewers) == 0 {
		return nil, domain.ErrNotEnoughMembers
	}

	return &reviewers[0], nil
}
