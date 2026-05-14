package eventhandler

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	eventservice "misis_kolhoz/internal/events/service"
	"misis_kolhoz/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	service *eventservice.Service
}

func NewHandler(s *eventservice.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) UploadData(contx context.Context) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		filePath := "internal/moked_data/events.xlsx"
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			filePath = "/app/" + filePath
		}

		err := h.service.LoadDataFromExcel(contx, filePath)
		if err != nil {
			logger.GetLoggerFromCtx(contx).Info(contx, "Failed load events data from excel", zap.Error(err))
			ctx.JSON(500, fmt.Sprintf("Failed load events data: %v", err))
			return
		}

		ctx.JSON(200, "Successfully loaded events data from excel")
	}
}

func (h *Handler) GetAll(contx context.Context) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		items, err := h.service.GetAllEvents(contx)
		if err != nil {
			logger.GetLoggerFromCtx(contx).Info(contx, "Failed get all events", zap.Error(err))
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"events": items})
	}
}

func (h *Handler) GetByMonth(contx context.Context) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		year, err := strconv.Atoi(ctx.Param("year"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid year"})
			return
		}
		month, err := strconv.Atoi(ctx.Param("month"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid month"})
			return
		}

		items, err := h.service.GetEventsByMonth(contx, year, month)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"events": items})
	}
}

func (h *Handler) GetUpcoming(contx context.Context) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		days := 30
		limit := 100

		if v := ctx.Query("days"); v != "" {
			parsed, err := strconv.Atoi(v)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid days query param"})
				return
			}
			days = parsed
		}

		if v := ctx.Query("limit"); v != "" {
			parsed, err := strconv.Atoi(v)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit query param"})
				return
			}
			limit = parsed
		}

		items, err := h.service.GetUpcomingEvents(contx, days, limit)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"events": items})
	}
}

func (h *Handler) GetByCategory(contx context.Context) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		category := ctx.Param("category")
		items, err := h.service.GetEventsByCategory(contx, category)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"events": items})
	}
}
