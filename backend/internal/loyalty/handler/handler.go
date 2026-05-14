package loyaltyhandler

import (
	"context"
	"fmt"
	"os"
	"strconv"

	loyaltyservice "misis_kolhoz/internal/loyalty/service"
	"misis_kolhoz/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	service *loyaltyservice.Service
}

func NewHandler(s *loyaltyservice.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) LoadOrders(contx context.Context) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		filePath := "internal/moked_data/orders.xlsx"
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			filePath = "/app/" + filePath
		}

		err := h.service.LoadOrdersFromExcel(contx, filePath)
		if err != nil {
			logger.GetLoggerFromCtx(contx).Info(contx, "Failed load orders from excel", zap.Error(err))
			ctx.JSON(500, fmt.Sprintf("Failed load orders: %v", err))
			return
		}

		ctx.JSON(200, "Successfully loaded orders and bonuses")
	}
}

func (h *Handler) GetAllClients(c *gin.Context) {
	clients, err := h.service.GetAllClients(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to get clients"})
		return
	}

	c.JSON(200, clients)
}

func (h *Handler) GetClientBonus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid client id"})
		return
	}

	clientInfo, err := h.service.GetClientBonusInfo(c.Request.Context(), id)
	if err != nil {
		c.JSON(404, gin.H{"error": "client not found"})
		return
	}

	c.JSON(200, clientInfo)
}

func (h *Handler) SpendBonus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid client id"})
		return
	}

	var req struct {
		Amount int `json:"amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	err = h.service.SpendClientBonus(c.Request.Context(), id, req.Amount)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "bonus spent successfully"})
}
