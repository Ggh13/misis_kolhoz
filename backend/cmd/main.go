package main

import (
	"context"
	"os"
	"os/signal"

	"misis_kolhoz/internal/config"
	farmerrepository "misis_kolhoz/internal/farmer/repository"
	farmerservice "misis_kolhoz/internal/farmer/service"
	farmerhandler "misis_kolhoz/internal/farmer/handler"
	"misis_kolhoz/internal/transport/rest"
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

	r, err := rest.NewRouter(ctx, cfg, farmerHandler)
	if err != nil {
		logger.GetLoggerFromCtx(ctx).Info(ctx, "Failed create router")
	}

	r.Run(ctx)

	<-ctx.Done()
	pgDB.Close()
	logger.GetLoggerFromCtx(ctx).Info(ctx, "Server Stopped")
}
