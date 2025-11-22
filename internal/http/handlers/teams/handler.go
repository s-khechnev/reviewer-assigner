package teams

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	"reviewer-assigner/internal/http/handlers"
	"reviewer-assigner/internal/logger"
	"reviewer-assigner/internal/service"
	"reviewer-assigner/internal/service/teams"
)

var validate = validator.New()

type TeamHandler struct {
	teamService *teams.TeamService
	log         *slog.Logger
}

func NewTeamHandler(log *slog.Logger, teamService *teams.TeamService) *TeamHandler {
	return &TeamHandler{
		log:         log,
		teamService: teamService,
	}
}

func (h *TeamHandler) AddTeam(c *gin.Context) {
	const op = "handlers.teams.AddTeam"
	log := h.log.With(slog.String("op", op))

	var req AddTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("failed to decode json body", logger.ErrAttr(err))

		c.JSON(http.StatusBadRequest, handlers.NewErrorResponse(handlers.ErrCodeInvalidJSON))
		return
	}

	log.Info("request decoded", slog.Any("request", req))

	if err := validate.Struct(req); err != nil {
		log.Warn("invalid json body", logger.ErrAttr(err))

		c.JSON(http.StatusUnprocessableEntity, handlers.NewErrorResponse(handlers.ErrCodeInvalidBody))
		return
	}

	newTeam, err := h.teamService.AddTeam(c.Copy(), req.TeamName, membersToDomain(req.Members))
	if errors.Is(err, service.ErrTeamAlreadyExists) {
		c.JSON(http.StatusBadRequest, handlers.NewErrorResponse(handlers.ErrCodeTeamExists, req.TeamName))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse(handlers.ErrCodeUnknown))
		return
	}

	c.JSON(http.StatusCreated, domainToAddTeamResponse(newTeam))
}

func (h *TeamHandler) GetTeam(c *gin.Context) {
	const op = "handlers.teams.GetTeam"
	log := h.log.With(slog.String("op", op))

	const teamNameParam = "team_name"

	teamName, ok := c.GetQuery(teamNameParam)
	if !ok {
		log.Warn(teamNameParam + " not found in query params")
		c.JSON(http.StatusBadRequest, handlers.NewErrorResponse(handlers.ErrCodeInvalidQueryParam))
		return
	}
	if teamName == "" {
		log.Warn(teamNameParam + " is empty")
		c.JSON(http.StatusBadRequest, handlers.NewErrorResponse(handlers.ErrCodeInvalidQueryParam))
		return
	}

	log.Info(teamNameParam+" param decoded", slog.Any(teamNameParam, teamName))

	team, err := h.teamService.GetTeam(c.Copy(), teamName)
	if errors.Is(err, service.ErrTeamNotFound) {
		c.JSON(http.StatusNotFound, handlers.NewErrorResponse(handlers.ErrCodeResourceNotFound))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse(handlers.ErrCodeUnknown))
		return
	}

	c.JSON(http.StatusOK, domainToGetTeamResponse(team))
}
