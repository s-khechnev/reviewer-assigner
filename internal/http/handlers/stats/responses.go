package stats

type GetStatsUserAssignmentsResponse struct {
	UserAssignments []UserAssignment `json:"assignments"`
}

type UserAssignment struct {
	UserID string `json:"user_id"`
	Name   string `json:"username"`
	Count  int    `json:"count"`
}
