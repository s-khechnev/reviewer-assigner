package service

import "errors"

var (
	ErrTeamAlreadyExists = errors.New("team already exists")
	ErrTeamNotFound      = errors.New("team not found")

	ErrUserNotFound = errors.New("user not found")
)
