package recommendationhandler

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	recommendationmodel "misis_kolhoz/internal/recommendation/model"
	recommendationservice "misis_kolhoz/internal/recommendation/service"
	"misis_kolhoz/pkg/logger"
)

type Service interface {
	GetCachedRecommendations(ctx context.Context, clientID int) (recommendationmodel.Response, error)
}

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) GetRecommendations(contx context.Context) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		idStr := ctx.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			logger.GetLoggerFromCtx(contx).Info(contx, "Failed parse client id", zap.Error(err))
			ctx.JSON(400, gin.H{"error": "invalid client id"})
			return
		}

		response, err := h.service.GetCachedRecommendations(ctx.Request.Context(), id)
		if err != nil {
			logger.GetLoggerFromCtx(contx).Info(contx, "Failed get recommendations", zap.Error(err))

			switch {
			case errors.Is(err, recommendationservice.ErrClientNotFound):
				ctx.JSON(404, gin.H{"error": "client not found"})
			case errors.Is(err, recommendationservice.ErrRecommendationsNotReady), errors.Is(err, recommendationservice.ErrCatalogNotLoaded):
				ctx.JSON(503, gin.H{"error": err.Error()})
			default:
				ctx.JSON(500, gin.H{"error": fmt.Sprintf("failed to get recommendations: %v", err)})
			}
			return
		}

		ctx.JSON(200, response)
	}
}
