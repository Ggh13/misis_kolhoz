package rest

import (
	"context"
	"fmt"
	"net/http"

	"misis_kolhoz/internal/config"
	"misis_kolhoz/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Router struct {
	httpServer *http.Server
}

func NewRouter(ctx context.Context, cfg *config.Config) (*Router, error) {
	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

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
