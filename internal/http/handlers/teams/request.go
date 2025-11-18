package teams

import teamsDomain "reviewer-assigner/internal/domain/teams"

type AddTeamRequest struct {
	TeamName string          `json:"team_name"`
	Members  []TeamMemberDTO `json:"members"`
}

type TeamMemberDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

func membersDTOtoDomain(members []TeamMemberDTO) []teamsDomain.Member {
	domainMembers := make([]teamsDomain.Member, 0, len(members))
	for _, member := range members {
		domainMembers = append(domainMembers, memberDTOtoDomain(&member))
	}
	return domainMembers
}

func memberDTOtoDomain(member *TeamMemberDTO) teamsDomain.Member {
	return teamsDomain.Member{
		Id:       member.UserID,
		Name:     member.Username,
		IsActive: member.IsActive,
	}
}
