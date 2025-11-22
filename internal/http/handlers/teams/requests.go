package teams

import teamsDomain "reviewer-assigner/internal/domain/teams"

type AddTeamRequest struct {
	TeamName string          `json:"team_name" validate:"required"`
	Members  []MemberRequest `json:"members"   validate:"required,dive"`
}

type MemberRequest struct {
	UserID   string `json:"user_id"   validate:"required"`
	Username string `json:"username"  validate:"required"`
	IsActive *bool  `json:"is_active" validate:"required"`
}

func membersToDomain(members []MemberRequest) []teamsDomain.Member {
	domainMembers := make([]teamsDomain.Member, 0, len(members))
	for _, member := range members {
		domainMembers = append(domainMembers, memberToDomain(&member))
	}
	return domainMembers
}

func memberToDomain(member *MemberRequest) teamsDomain.Member {
	return teamsDomain.Member{
		ID:       member.UserID,
		Name:     member.Username,
		IsActive: *member.IsActive,
	}
}
