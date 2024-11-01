package logger

import (
	"errors"

	"go.uber.org/zap"
)

var Log *zap.Logger
var ErrInitLogger = errors.New("failed to initialize logger")

func Init() {
	var err error
	Log, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}
}

// Usage example:
// logger.Log.Info("Processing request",
//     zap.String("remote_addr", r.RemoteAddr),
//     zap.Duration("latency", latency),
// )
