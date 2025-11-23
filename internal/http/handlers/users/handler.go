package users

import (
	"errors"
	"log/slog"
	"net/http"
	"reviewer-assigner/internal/http/handlers"
	"reviewer-assigner/internal/logger"
	"reviewer-assigner/internal/service"
	"reviewer-assigner/internal/service/users"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

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
	const op = "handlers.users.UpdateIsActive"
	log := h.log.With(slog.String("op", op))

	var req SetIsActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("invalid json body", logger.ErrAttr(err))

		c.JSON(http.StatusBadRequest, handlers.NewErrorResponse(handlers.ErrCodeInvalidJSON))
		return
	}

	log.Info("request decoded", slog.Any("request", req))

	if err := validate.Struct(req); err != nil {
		log.Warn("validation error", logger.ErrAttr(err))

		c.JSON(
			http.StatusUnprocessableEntity,
			handlers.NewErrorResponse(handlers.ErrCodeInvalidBody),
		)
		return
	}

	user, err := h.userService.SetIsActive(c.Request.Context(), req.UserID, *req.IsActive)
	if errors.Is(err, service.ErrUserNotFound) {
		c.JSON(http.StatusNotFound, handlers.NewErrorResponse(handlers.ErrCodeResourceNotFound))
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse(handlers.ErrCodeUnknown))
		return
	}

	c.JSON(http.StatusOK, domainToSetIsActiveResponse(user))
}

func (h *UserHandler) GetReview(c *gin.Context) {
	const op = "handlers.users.GetReview"
	log := h.log.With(slog.String("op", op))

	const userIDParam = "user_id"

	userID, ok := c.GetQuery(userIDParam)
	if !ok {
		log.Warn(userIDParam + " not found in query params")

		c.JSON(http.StatusBadRequest, handlers.NewErrorResponse(handlers.ErrCodeInvalidQueryParam))
		return
	}
	if userID == "" {
		log.Warn(userIDParam + " is empty")

		c.JSON(http.StatusBadRequest, handlers.NewErrorResponse(handlers.ErrCodeInvalidQueryParam))
		return
	}

	log.Info(userIDParam+" param decoded", slog.Any(userIDParam, userID))

	pullRequests, err := h.userService.GetReview(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse(handlers.ErrCodeUnknown))
		return
	}

	c.JSON(http.StatusOK, domainToGetReviewResponse(userID, pullRequests))
}
