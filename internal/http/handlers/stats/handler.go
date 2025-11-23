package stats

import (
	"context"
	"log/slog"
	"net/http"
	prsDomain "reviewer-assigner/internal/domain/pullrequests"
	"reviewer-assigner/internal/http/handlers"
	"reviewer-assigner/internal/logger"
	"strings"

	"github.com/gin-gonic/gin"
)

type StatRepository interface {
	GetStatsReviewersAssignments(
		ctx context.Context,
		status string,
		activeOnly bool,
	) ([]UserAssignment, error)
}

type StatHandler struct {
	log       *slog.Logger
	statsRepo StatRepository
}

func NewStatHandler(log *slog.Logger, statsRepo StatRepository) *StatHandler {
	return &StatHandler{
		log:       log,
		statsRepo: statsRepo,
	}
}

func (h *StatHandler) GetStatsReviewersAssignments(c *gin.Context) {
	const op = "handlers.stats.GetStatsReviewersAssignments"
	log := h.log.With(slog.String("op", op))

	const statusParam = "status"
	const activeOnlyParam = "active_only"

	status := strings.ToUpper(c.Query(statusParam))
	if !isValidStatus(status) {
		log.Warn("invalid status", slog.String("status", status))

		c.JSON(http.StatusBadRequest, handlers.NewErrorResponse(handlers.ErrCodeInvalidQueryParam))
		return
	}

	_, ok := c.GetQuery(activeOnlyParam)
	activeOnly := ok

	log.Info(
		"query param decoded",
		slog.String(statusParam, status),
		slog.Bool(activeOnlyParam, activeOnly),
	)

	usersAssignments, err := h.statsRepo.GetStatsReviewersAssignments(
		c.Request.Context(),
		status,
		activeOnly,
	)
	if err != nil {
		log.Error("failed to get stats reviewer assignments", logger.ErrAttr(err))

		c.JSON(http.StatusInternalServerError, handlers.NewErrorResponse(handlers.ErrCodeUnknown))
		return
	}

	log.Info("got stats reviewer assignments", slog.Int("len", len(usersAssignments)))

	c.JSON(http.StatusOK, GetStatsUserAssignmentsResponse{
		UserAssignments: usersAssignments,
	})
}

func isValidStatus(status string) bool {
	// can be empty, it means all statuses
	if status == "" {
		return true
	}

	return status == string(prsDomain.StatusOpen) ||
		status == string(prsDomain.StatusMerged)
}
