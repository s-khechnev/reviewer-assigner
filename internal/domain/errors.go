package domain

import "errors"

var (
	ErrNotEnoughMembers = errors.New("not enough members")

	ErrTeamMembersMismatch = errors.New("members mismatch")

	ErrPullRequestAlreadyMerged = errors.New("pull request already merged")
)
