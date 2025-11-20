package service

import "errors"

var (
	ErrTeamAlreadyExists = errors.New("team already exists")
	ErrTeamNotFound      = errors.New("team not found")

	ErrUserNotFound = errors.New("user not found")

	ErrPullRequestAlreadyExists = errors.New("pull request already exists")
	ErrPullRequestNotFound      = errors.New("pull request not found")
	ErrPullRequestAlreadyMerged = errors.New("pull request already merged")
	ErrPullRequestNotAssigned   = errors.New("reviewer is not assigned to this PR")
	ErrPullRequestNoCandidates  = errors.New("no active replacement candidate in team")
)
