package eventrouter

import (
	"context"

	"github.com/gin-gonic/gin"
	eventhandler "misis_kolhoz/internal/events/handler"
)

func Transport(r *gin.Engine, h *eventhandler.Handler, ctx context.Context) {
	r.POST("/upload_events", h.UploadData(ctx))
	r.GET("/events", h.GetAll(ctx))
	r.GET("/events/month/:year/:month", h.GetByMonth(ctx))
	r.GET("/events/upcoming", h.GetUpcoming(ctx))
	r.GET("/events/category/:category", h.GetByCategory(ctx))
}
