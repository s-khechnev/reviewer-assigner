package pull_requests

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	"reviewer-assigner/internal/http/handlers"
	"reviewer-assigner/internal/logger"
	"reviewer-assigner/internal/service"
	prs "reviewer-assigner/internal/service/pull_requests"
)

var validate = validator.New()

type PullRequestHandler struct {
	pullRequestService *prs.PullRequestService
	log                *slog.Logger
}

func NewPullRequestHandler(log *slog.Logger, pullRequestService *prs.PullRequestService) *PullRequestHandler {
	return &PullRequestHandler{
		pullRequestService: pullRequestService,
		log:                log,
	}
}

func (h *PullRequestHandler) Create(c *gin.Context) {
	const op = "handlers.pull_requests.Create"
	log := h.log.With(slog.String("op", op))

	var req CreatePullRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("invalid json body", logger.ErrAttr(err))

		c.JSON(http.StatusBadRequest, handlers.NewErrorResponse(handlers.ErrCodeInvalidJSON))
		return
	}

	log.Info("request decoded", slog.Any("request", req))

	if err := validate.Struct(req); err != nil {
		log.Warn("validation error", logger.ErrAttr(err))

		c.JSON(http.StatusUnprocessableEntity, handlers.NewErrorResponse(handlers.ErrCodeInvalidBody))
		return
	}

	pullRequest, err := h.pullRequestService.Create(c.Copy(), req.ID, req.Name, req.AuthorID)
	if errors.Is(err, service.ErrUserNotFound) || errors.Is(err, service.ErrTeamNotFound) {
		c.JSON(http.StatusNotFound, handlers.NewErrorResponse(handlers.ErrCodeResourceNotFound))
		return
	}
	if errors.Is(err, service.ErrPullRequestAlreadyExists) {
		c.JSON(http.StatusConflict, handlers.NewErrorResponse(handlers.ErrCodePullRequestExists, req.ID))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse(handlers.ErrCodeUnknown))
		return
	}

	c.JSON(http.StatusCreated, domainToCreatePullRequestResponse(pullRequest))
}

func (h *PullRequestHandler) Merge(c *gin.Context) {
	const op = "handlers.pull_requests.Merge"
	log := h.log.With(slog.String("op", op))

	var req MergePullRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("invalid json body", logger.ErrAttr(err))

		c.JSON(http.StatusBadRequest, handlers.NewErrorResponse(handlers.ErrCodeInvalidJSON))
		return
	}

	log.Info("request decoded", slog.Any("request", req))

	pullRequest, err := h.pullRequestService.Merge(c.Copy(), req.ID)
	if errors.Is(err, service.ErrPullRequestNotFound) {
		c.JSON(http.StatusNotFound, handlers.NewErrorResponse(handlers.ErrCodeResourceNotFound))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse(handlers.ErrCodeUnknown))
		return
	}

	c.JSON(http.StatusOK, domainToMergePullRequestResponse(pullRequest))
}

func (h *PullRequestHandler) Reassign(c *gin.Context) {
	const op = "handlers.pull_requests.Reassign"
	log := h.log.With(slog.String("op", op))

	var req ReassignPullRequestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("invalid json body", logger.ErrAttr(err))

		c.JSON(http.StatusBadRequest, handlers.NewErrorResponse(handlers.ErrCodeInvalidJSON))
		return
	}

	log.Info("request decoded", slog.Any("request", req))

	pullRequest, replacedBy, err := h.pullRequestService.Reassign(c.Copy(), req.ID, req.OldReviewerID)
	if errors.Is(err, service.ErrPullRequestNotFound) {
		c.JSON(http.StatusBadRequest, handlers.NewErrorResponse(handlers.ErrCodeResourceNotFound))
		return
	}
	if errors.Is(err, service.ErrPullRequestAlreadyMerged) {
		c.JSON(http.StatusConflict, handlers.NewErrorResponse(handlers.ErrCodePullRequestMerged))
		return
	}
	if errors.Is(err, service.ErrPullRequestNotAssigned) {
		c.JSON(http.StatusConflict, handlers.NewErrorResponse(handlers.ErrCodePullRequestNotAssigned))
		return
	}
	if errors.Is(err, service.ErrPullRequestNoCandidates) {
		c.JSON(http.StatusBadRequest, handlers.NewErrorResponse(handlers.ErrCodePullRequestNoCandidates))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse(handlers.ErrCodeUnknown))
		return
	}

	c.JSON(http.StatusOK, domainToReassignPullRequestResponse(pullRequest, replacedBy))
}
