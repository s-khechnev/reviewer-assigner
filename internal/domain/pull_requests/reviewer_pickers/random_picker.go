package reviewer_pickers

import (
	"math/rand"
	teamsDomain "reviewer-assigner/internal/domain/teams"
)

type RandomReviewerPicker struct{}

func (p *RandomReviewerPicker) Pick(members []teamsDomain.Member, count int) []teamsDomain.Member {
	if len(members) == 0 || count <= 0 {
		return nil
	}

	if len(members) <= count {
		return members
	}

	rand.Shuffle(len(members), func(i, j int) {
		members[i], members[j] = members[j], members[i]
	})

	reviewers := make([]teamsDomain.Member, count)
	copy(reviewers, members[:count])

	return reviewers
}
