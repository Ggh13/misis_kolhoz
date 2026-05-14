package farmerhandler

import (
	"context"
	"fmt"
	"os"
	"strconv"

	farmerservice "misis_kolhoz/internal/farmer/service"
	"misis_kolhoz/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	service *farmerservice.Service
}

func NewHandler(s *farmerservice.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) UploadData(contx context.Context) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		filePath := "internal/moked_data/farmers_sku.xlsx"
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			filePath = "/app/" + filePath
		}

		err := h.service.LoadDataFromExcel(contx, filePath)
		if err != nil {
			logger.GetLoggerFromCtx(contx).Info(contx, "Failed load data from excel", zap.Error(err))
			ctx.JSON(500, fmt.Sprintf("Failed load data: %v", err))
			return
		}

		ctx.JSON(200, "Successfully loaded data from excel. Run reseed script for embeddings.")
	}
}

func (h *Handler) GetFarmerData(contx context.Context) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		idStr := ctx.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			logger.GetLoggerFromCtx(contx).Info(contx, "Failed parse farmer id", zap.Error(err))
			ctx.JSON(400, "Invalid farmer ID")
			return
		}

		farmer, err := h.service.GetFarmerByID(contx, id)
		if err != nil {
			logger.GetLoggerFromCtx(contx).Info(contx, "Failed get farmer data", zap.Error(err))
			ctx.JSON(404, "Farmer not found")
			return
		}

		ctx.JSON(200, farmer)
	}
}
