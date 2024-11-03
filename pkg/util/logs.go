package util

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

// LogFromContext returns a logr.Logger from the context or an instance of the global logger
func LogFromContext(ctx context.Context) logr.Logger {
	l, err := logr.FromContext(ctx)
	if err != nil {
		return zapr.NewLogger(zap.L())
	}
	return l
}
