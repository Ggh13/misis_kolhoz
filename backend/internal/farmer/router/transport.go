package farmerrouter

import (
	"context"

	"github.com/gin-gonic/gin"
	farmerhandler "misis_kolhoz/internal/farmer/handler"
)

func Transport(r *gin.Engine, h *farmerhandler.Handler, ctx context.Context) {
	r.POST("/upload_data", h.UploadData(ctx))
	r.GET("/farmer_data/:id", h.GetFarmerData(ctx))
	r.GET("/farmers/search", h.SearchFarmers(ctx))
}