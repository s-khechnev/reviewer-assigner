package users

import (
	"errors"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"reviewer-assigner/internal/http/handlers"
	"reviewer-assigner/internal/logger"
	"reviewer-assigner/internal/service"
	"reviewer-assigner/internal/service/users"
)

type UserHandler struct {
	userService *users.UserService
	log         *slog.Logger
}

func NewUserHandler(log *slog.Logger, userService *users.UserService) *UserHandler {
	return &UserHandler{
		log:         log,
		userService: userService,
	}
}

func (h *UserHandler) SetIsActive(c *gin.Context) {
	const op = "handlers.users.SetIsActive"
	log := h.log.With(slog.String("op", op))

	var req SetIsActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("invalid json body", logger.ErrAttr(err))

		c.JSON(http.StatusBadRequest, handlers.NewErrorResponse(handlers.ErrCodeInvalidJSON))
		return
	}

	log.Info("request decoded", slog.Any("request", req))

	user, err := h.userService.SetIsActive(c.Copy(), req.UserID, req.IsActive)
	if errors.Is(err, service.ErrUserNotFound) {
		c.JSON(http.StatusBadRequest, handlers.NewErrorResponse(handlers.ErrCodeResourceNotFound))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse(handlers.ErrCodeUnknown))
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user": domainToUserResponse(user)})
}

func (h *UserHandler) GetReview(c *gin.Context) {
	const op = "handlers.users.GetReview"
	log := h.log.With(slog.String("op", op))

	const UserIDParam = "user_id"

	userID, ok := c.GetQuery(UserIDParam)
	if !ok {
		log.Warn(UserIDParam + " not found in query params")
		c.JSON(http.StatusBadRequest, handlers.NewErrorResponse(handlers.ErrCodeInvalidQueryParam))
		return
	}
	if userID == "" {
		log.Warn(UserIDParam + " is empty")
		c.JSON(http.StatusBadRequest, handlers.NewErrorResponse(handlers.ErrCodeInvalidQueryParam))
		return
	}

	log.Info(UserIDParam+" param decoded", slog.Any(UserIDParam, userID))

	pullRequests, err := h.userService.GetReview(c.Copy(), userID)
	if errors.Is(err, service.ErrTeamNotFound) {
		c.JSON(http.StatusNotFound, handlers.NewErrorResponse(handlers.ErrCodeResourceNotFound))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse(handlers.ErrCodeUnknown))
		return
	}

	c.JSON(http.StatusOK, domainToGetReviewResponse(userID, pullRequests))
}
