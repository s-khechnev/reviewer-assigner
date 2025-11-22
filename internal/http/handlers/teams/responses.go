package teams

import teamsDomain "reviewer-assigner/internal/domain/teams"

type AddTeamResponse struct {
	TeamResponse `json:"team"`
}

type GetTeamResponse struct {
	TeamResponse
}

type TeamResponse struct {
	TeamName string           `json:"team_name"`
	Members  []MemberResponse `json:"members"`
}

type MemberResponse struct {
	ID       string `json:"user_id"`
	Name     string `json:"username"`
	IsActive bool   `json:"is_active"`
}

func domainToAddTeamResponse(team *teamsDomain.Team) *AddTeamResponse {
	return &AddTeamResponse{
		*domainToTeamResponse(team),
	}
}

func domainToGetTeamResponse(team *teamsDomain.Team) *GetTeamResponse {
	return &GetTeamResponse{
		*domainToTeamResponse(team),
	}
}

func domainToTeamResponse(team *teamsDomain.Team) *TeamResponse {
	members := make([]MemberResponse, 0, len(team.Members))
	for _, member := range team.Members {
		members = append(members, MemberResponse{
			ID:       member.ID,
			Name:     member.Name,
			IsActive: member.IsActive,
		})
	}

	return &TeamResponse{
		TeamName: team.Name,
		Members:  members,
	}
}
