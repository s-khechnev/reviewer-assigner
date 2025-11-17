package reviewer_pickers

import (
	"math/rand"
	teamsDomain "reviewer-assigner/internal/domain/teams"
)

type RandomReviewerPicker struct {
	countReviewers uint
}

func NewRandomReviewerPicker(count uint) *RandomReviewerPicker {
	return &RandomReviewerPicker{countReviewers: count}
}

func (p *RandomReviewerPicker) Pick(members []teamsDomain.Member) ([]teamsDomain.Member, error) {
	if uint(len(members)) < p.countReviewers {
		return nil, ErrNotEnoughMembers
	}

	if uint(len(members)) == p.countReviewers {
		return members, nil
	}

	rand.Shuffle(len(members), func(i, j int) {
		members[i], members[j] = members[j], members[i]
	})

	reviewers := make([]teamsDomain.Member, p.countReviewers)
	copy(reviewers, members[:p.countReviewers])

	return reviewers, nil
}
