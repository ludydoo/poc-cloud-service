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
	if ctx == nil {
		return logger
	}
	if tenant := GetTenant(ctx); tenant != "" {
		return logger.With(zap.String("tenant", tenant))
	}
	return logger
}

type tenantKey struct{}

func WithTenant(ctx context.Context, tenant string) context.Context {
	return context.WithValue(ctx, tenantKey{}, tenant)
}

func GetTenant(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if tenant, ok := ctx.Value(tenantKey{}).(string); ok {
		return tenant
	}
	return ""
}
