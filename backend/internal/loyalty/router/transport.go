package loyaltyrouter

import (
	"context"

	"github.com/gin-gonic/gin"
	loyaltyhandler "misis_kolhoz/internal/loyalty/handler"
)

func Transport(r *gin.Engine, h *loyaltyhandler.Handler, ctx context.Context) {
	r.POST("/load_orders", h.LoadOrders(ctx))
	r.GET("/clients", h.GetAllClients)
	r.GET("/client_bonus/:id", h.GetClientBonus)
	r.POST("/spend_bonus/:id", h.SpendBonus)
}
