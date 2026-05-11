package main

import (
	"context"
	"os"
	"os/signal"

	"misis_kolhoz/internal/config"
	farmerhandler "misis_kolhoz/internal/farmer/handler"
	farmerrepository "misis_kolhoz/internal/farmer/repository"
	farmerservice "misis_kolhoz/internal/farmer/service"
	loyaltyhandler "misis_kolhoz/internal/loyalty/handler"
	loyaltyrepository "misis_kolhoz/internal/loyalty/repository"
	loyaltyservice "misis_kolhoz/internal/loyalty/service"
	"misis_kolhoz/internal/transport/rest"
	vectorhandler "misis_kolhoz/internal/vector/handler"
	vectorrepository "misis_kolhoz/internal/vector/repository"
	vectorservice "misis_kolhoz/internal/vector/service"
	"misis_kolhoz/pkg/logger"
	"misis_kolhoz/pkg/postgres"
	"misis_kolhoz/pkg/qdrant"

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
	logger.GetLoggerFromCtx(ctx).Info(ctx, "Succesfully load config")

	pgDB, err := postgres.NewPostgres(ctx, &cfg.PostgresCFG)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "Failsed connect to postgres DB", zap.Error(err))
	}
	if err := pgDB.Ping(ctx); err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "Failed ping pgDB", zap.Error(err))
	}
	logger.GetLoggerFromCtx(ctx).Info(ctx, "Succesfully connected to pgDB")

	farmerRepo := farmerrepository.NewRepository(pgDB)
	farmerService := farmerservice.NewService(farmerRepo)
	farmerHandler := farmerhandler.NewHandler(farmerService)

	qdrantClient, err := qdrant.NewQdrant(ctx, &qdrant.Config{
		Host:           cfg.QdrantCFG.Host,
		Port:           cfg.QdrantCFG.Port,
		CollectionName: cfg.QdrantCFG.CollectionName,
		VectorSize:     cfg.QdrantCFG.VectorSize,
	})
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Fatal(ctx, "Failed connect to qdrant", zap.Error(err))
	}
	logger.GetLoggerFromCtx(ctx).Info(ctx, "Succesfully connected to qdrant")

	vectorRepo := vectorrepository.NewVectorRepository(qdrantClient, cfg.QdrantCFG.CollectionName)
	vectorService := vectorservice.NewVectorService(vectorRepo)
	vectorHandler := vectorhandler.NewVectorHandler(vectorService)

	loyaltyRepo := loyaltyrepository.NewRepository(pgDB)
	loyaltyService := loyaltyservice.NewService(loyaltyRepo)
	loyaltyHandler := loyaltyhandler.NewHandler(loyaltyService)

	r, err := rest.NewRouter(ctx, cfg, farmerHandler, vectorHandler, loyaltyHandler)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "Failed create router")
	}

	r.Run(ctx)

	<-ctx.Done()
	pgDB.Close()
	logger.GetLoggerFromCtx(ctx).Info(ctx, "Server Stopped")
}
