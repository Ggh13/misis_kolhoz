package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"misis_kolhoz/internal/config"
	farmerhandler "misis_kolhoz/internal/farmer/handler"
	farmerrepository "misis_kolhoz/internal/farmer/repository"
	farmerservice "misis_kolhoz/internal/farmer/service"
	loyaltyhandler "misis_kolhoz/internal/loyalty/handler"
	loyaltyrepository "misis_kolhoz/internal/loyalty/repository"
	loyaltyservice "misis_kolhoz/internal/loyalty/service"
	recommendationhandler "misis_kolhoz/internal/recommendation/handler"
	recommendationrepository "misis_kolhoz/internal/recommendation/repository"
	recommendationservice "misis_kolhoz/internal/recommendation/service"
	"misis_kolhoz/internal/transport/rest"
	vectorhandler "misis_kolhoz/internal/vector/handler"
	vectorrepository "misis_kolhoz/internal/vector/repository"
	vectorservice "misis_kolhoz/internal/vector/service"
	"misis_kolhoz/pkg/logger"
	"misis_kolhoz/pkg/postgres"

	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	ctx, _ = logger.NewLogger(ctx)

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	cfg, err := config.NewConfig(ctx)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "Failed load config", zap.Error(err))
	}
	logger.GetLoggerFromCtx(ctx).Info(ctx, "Successfully load config")

	pgDB, err := postgres.NewPostgres(ctx, &cfg.PostgresCFG)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "Failed connect to postgres DB", zap.Error(err))
	}
	if err := pgDB.Ping(ctx); err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "Failed ping pgDB", zap.Error(err))
	}
	logger.GetLoggerFromCtx(ctx).Info(ctx, "Successfully connected to pgDB")

	// Init vector table
	vectorRepo := vectorrepository.NewVectorRepository(pgDB)
	if err := vectorRepo.Init(ctx); err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "Failed init vector table", zap.Error(err))
	}
	logger.GetLoggerFromCtx(ctx).Info(ctx, "Vector table initialized")

	farmerRepo := farmerrepository.NewRepository(pgDB)
	farmerService := farmerservice.NewService(farmerRepo)
	farmerHandler := farmerhandler.NewHandler(farmerService)

	recommendationRepo := recommendationrepository.NewRepository(pgDB)
	recommendationService := recommendationservice.NewService(recommendationRepo)
	ordersFilePath := resolveLocalOrDockerPath("internal/moked_data/orders.xlsx")
	if err := recommendationService.StartPeriodicRefresh(ctx, ordersFilePath, time.Minute); err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "Recommendations will retry refresh in background", zap.Error(err), zap.String("filePath", ordersFilePath))
	}
	recommendationHandler := recommendationhandler.NewHandler(recommendationService)

	vectorService := vectorservice.NewVectorService(vectorRepo)
	vectorHandler := vectorhandler.NewVectorHandler(vectorService)

	loyaltyRepo := loyaltyrepository.NewRepository(pgDB)
	loyaltyService := loyaltyservice.NewService(loyaltyRepo)
	loyaltyHandler := loyaltyhandler.NewHandler(loyaltyService)

	r, err := rest.NewRouter(ctx, cfg, farmerHandler, vectorHandler, loyaltyHandler, recommendationHandler)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "Failed create router")
	}

	r.Run(ctx)

	<-ctx.Done()
	pgDB.Close()
	logger.GetLoggerFromCtx(ctx).Info(ctx, "Server Stopped")
}

func resolveLocalOrDockerPath(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "/app/" + path
	}

	return path
}
