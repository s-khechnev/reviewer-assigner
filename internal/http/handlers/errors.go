package handlers

import (
	"fmt"
)

type ErrCode string

const (
	ErrCodeInvalidJSON       ErrCode = "INVALID_JSON"
	ErrCodeInvalidQueryParam ErrCode = "INVALID_QUERY_PARAM"
	ErrCodeInvalidBody       ErrCode = "INVALID_BODY"

	ErrCodeTeamExists ErrCode = "TEAM_EXISTS"

	ErrCodePullRequestExists      ErrCode = "PR_EXISTS"
	ErrCodePullRequestMerged      ErrCode = "PR_MERGED"
	ErrCodePullRequestNotAssigned ErrCode = "NOT_ASSIGNED"
	ErrCodePullRequestNoCandidate ErrCode = "NO_CANDIDATE"

	ErrCodeResourceNotFound ErrCode = "NOT_FOUND"

	ErrCodeUnknown ErrCode = "UNKNOWN"
)

var errCodeMessages = map[ErrCode]string{
	ErrCodeInvalidJSON:       "invalid JSON format",
	ErrCodeInvalidQueryParam: "invalid query parameter",
	ErrCodeInvalidBody:       "invalid request body",

	ErrCodeTeamExists: "%s already exists",

	ErrCodePullRequestExists:      "PR %s already exists",
	ErrCodePullRequestMerged:      "cannot reassign on merged PR",
	ErrCodePullRequestNotAssigned: "reviewer is not assigned to this PR",
	ErrCodePullRequestNoCandidate: "no active replacement candidate in team",

	ErrCodeResourceNotFound: "resource not found",
}

type ErrorResponse struct {
	Error ErrorDetails `json:"error"`
}

type ErrorDetails struct {
	Code    ErrCode `json:"code"`
	Message string  `json:"message"`
}

func (e ErrCode) Message(args ...any) string {
	if msg, ok := errCodeMessages[e]; ok {
		return fmt.Sprintf(msg, args...)
	}

	return "Unknown error"
}

func NewErrorResponse(code ErrCode, args ...any) *ErrorResponse {
	return &ErrorResponse{
		Error: ErrorDetails{
			Code:    code,
			Message: code.Message(args...),
		},
	}
}
