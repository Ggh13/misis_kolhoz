package rest

import (
	"context"
	"fmt"
	"net/http"

	"misis_kolhoz/internal/config"
	farmerhandler "misis_kolhoz/internal/farmer/handler"
	farmerrouter "misis_kolhoz/internal/farmer/router"
	loyaltyhandler "misis_kolhoz/internal/loyalty/handler"
	loyaltyrouter "misis_kolhoz/internal/loyalty/router"
	vectorhandler "misis_kolhoz/internal/vector/handler"
	vectorrouter "misis_kolhoz/internal/vector/router"
	"misis_kolhoz/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Router struct {
	httpServer *http.Server
}

func NewRouter(ctx context.Context, cfg *config.Config, farmerH *farmerhandler.Handler, vectorH *vectorhandler.VectorHandler, loyaltyH *loyaltyhandler.Handler) (*Router, error) {
	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	farmerrouter.Transport(router, farmerH, ctx)
	vectorrouter.NewRouter(router, vectorH)
	loyaltyrouter.Transport(router, loyaltyH, ctx)

	return &Router{
		httpServer: &http.Server{
			Addr:    fmt.Sprintf("%s:%s", cfg.RestHost, cfg.RestPort),
			Handler: router,
		},
	}, nil
}

func (r *Router) Run(ctx context.Context) {
	logger.GetLoggerFromCtx(ctx).Info(ctx, "Starting server", zap.String("addr", r.httpServer.Addr))

	go func() {
		if err := r.httpServer.ListenAndServe(); err != nil {
			logger.GetLoggerFromCtx(ctx).Fatal(ctx, "Server error", zap.Error(err))
		}
	}()
}
