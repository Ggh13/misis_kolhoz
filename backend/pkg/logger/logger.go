package logger

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type key string

const (
	KeyLogger = key("logger")

	RequestID = key("request_id")
)

type Logger struct {
	logger *zap.Logger
}

func NewLogger(ctx context.Context) (context.Context, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failsed create logger logger.Newlogger: %w", err)
	}

	ctx = context.WithValue(ctx, KeyLogger, &Logger{logger: logger})

	return ctx, nil
}

func GetLoggerFromCtx(ctx context.Context) *Logger {
	return ctx.Value(KeyLogger).(*Logger)
}

func (l *Logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx.Value(RequestID) != nil {
		fields = append(fields, zap.String(string(RequestID), ctx.Value(RequestID).(string)))
	}
	l.logger.Info(msg, fields...)
}

func (l *Logger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	if ctx.Value(RequestID) != nil {
		fields = append(fields, zap.String(string(RequestID), ctx.Value(RequestID).(string)))
	}
	l.logger.Fatal(msg, fields...)
}

func MiddleWare(contx context.Context, logger *Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		guid := uuid.New().String()
		contx = context.WithValue(contx, RequestID, guid)
		contx = context.WithValue(contx, KeyLogger, logger)

		logger := GetLoggerFromCtx(contx)

		logger.Info(ctx,
			"request", zap.String("restURL", ctx.FullPath()),
			zap.Time("request time", time.Now()),
		)

	}
}
