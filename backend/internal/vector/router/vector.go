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
		vector.POST("/distance", h.Distance)
	}
}