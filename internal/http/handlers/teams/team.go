package teams

import (
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	service "reviewer-assigner/internal/service/teams"
)

type TeamHandler struct {
	teamService *service.TeamService
	log         *slog.Logger
}

func NewTeamHandler(log *slog.Logger, teamService *service.TeamService) *TeamHandler {
	return &TeamHandler{
		log:         log,
		teamService: teamService,
	}
}

func (h *TeamHandler) AddTeam(c *gin.Context) {
	var req AddTeamRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, "123")
		return
	}

	id, err := h.teamService.AddTeam(c.Copy(), req.TeamName, membersDTOtoDomain(req.Members))
	if err != nil {
		c.JSON(http.StatusInternalServerError, "123")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func (h *TeamHandler) GetTeam(c *gin.Context) {

}
