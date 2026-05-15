package router

import (
	"github.com/gin-gonic/gin"
	"misis_kolhoz/internal/vector/handler"
)

func NewRouter(r *gin.Engine, h *handler.VectorHandler) {
	vector := r.Group("/vector")
	{
		vector.POST("/:id", h.Create)
		vector.GET("/:id", h.Get)
		vector.PUT("/:id", h.Update)
		vector.DELETE("/:id", h.Delete)

		vector.POST("/search", h.Search)
		vector.POST("/events/search", h.SearchEvents)
		vector.POST("/events/for-product", h.SearchEventsForProduct)
		vector.POST("/distance", h.Distance)
	}
}