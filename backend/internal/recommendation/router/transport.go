package recommendationrouter

import (
	"context"

	"github.com/gin-gonic/gin"
	recommendationhandler "misis_kolhoz/internal/recommendation/handler"
)

func Transport(r *gin.Engine, h *recommendationhandler.Handler, ctx context.Context) {
	r.GET("/recommendations/:id", h.GetRecommendations(ctx))
}
