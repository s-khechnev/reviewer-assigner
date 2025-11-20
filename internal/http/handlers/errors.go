package handlers

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

type ErrCode string

const (
	ErrCodeInvalidJSON       ErrCode = "INVALID_JSON"
	ErrCodeInvalidQueryParam ErrCode = "INVALID_QUERY_PARAM"

	ErrCodeTeamExists ErrCode = "TEAM_EXISTS"

	ErrCodePullRequestExists       ErrCode = "PR_EXISTS"
	ErrCodePullRequestMerged       ErrCode = "PR_MERGED"
	ErrCodePullRequestNotAssigned  ErrCode = "NOT_ASSIGNED"
	ErrCodePullRequestNoCandidates ErrCode = "NO_CANDIDATES"

	ErrCodeResourceNotFound ErrCode = "NOT_FOUND"

	ErrCodeUnknown ErrCode = "UNKNOWN"
)

var errCodeMessages = map[ErrCode]string{
	ErrCodeInvalidJSON:       "invalid JSON format",
	ErrCodeInvalidQueryParam: "invalid query parameter",

	ErrCodeTeamExists: "%s already exists",

	ErrCodePullRequestExists:       "PR %s already exists",
	ErrCodePullRequestMerged:       "cannot reassign on merged PR",
	ErrCodePullRequestNotAssigned:  "reviewer is not assigned to this PR",
	ErrCodePullRequestNoCandidates: "no active replacement candidate in team",

	ErrCodeResourceNotFound: "resource not found",
}

func (e ErrCode) Message(args ...any) string {
	if msg, ok := errCodeMessages[e]; ok {
		return fmt.Sprintf(msg, args...)
	}

	return "Unknown error"
}

func NewErrorResponse(code ErrCode, args ...any) gin.H {
	return gin.H{
		"error": gin.H{
			"code":    code,
			"message": code.Message(args...),
		},
	}
}
