package log

import (
	"context"
	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	logger, _ = zap.NewProduction()
}

func FromContext(ctx context.Context) *zap.Logger {
	return logger
}
