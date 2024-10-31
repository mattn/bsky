package pkg

import (
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
	BuiltBy = "unknown"
)

func LogVersion() {
	log := zapr.NewLogger(zap.L())
	log.Info("binary version", "version", Version, "commit", Commit, "date", Date, "builtBy", BuiltBy)
}
