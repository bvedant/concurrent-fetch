package logger

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestLogger(t *testing.T) {
	// Create an observer core for testing
	observedZapCore, observedLogs := observer.New(zapcore.DebugLevel)
	originalLogger := Log
	Log = zap.New(observedZapCore)
	// Restore the original logger after the test
	defer func() { Log = originalLogger }()

	tests := []struct {
		name       string
		logFunc    func()
		level      zapcore.Level
		message    string
		checkField func([]zap.Field) bool
	}{
		{
			name: "error_level",
			logFunc: func() {
				Log.Error("test error message", zap.Error(ErrInitLogger))
			},
			level:   zapcore.ErrorLevel,
			message: "test error message",
			checkField: func(fields []zap.Field) bool {
				for _, f := range fields {
					if f.Key == "error" {
						return true
					}
				}
				return false
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			observedLogs.TakeAll()
			tt.logFunc()

			logs := observedLogs.All()
			if len(logs) != 1 {
				t.Errorf("Expected 1 log entry, got %d", len(logs))
				return
			}

			log := logs[0]
			if log.Level != tt.level {
				t.Errorf("Expected level %v, got %v", tt.level, log.Level)
			}
			if log.Message != tt.message {
				t.Errorf("Expected message %q, got %q", tt.message, log.Message)
			}

			fields := log.Context
			if !tt.checkField(fields) {
				t.Errorf("Expected field check to pass, but it didn't")
			}
		})
	}
}

func TestLoggerInit(t *testing.T) {
	// Save original logger
	originalLogger := Log
	defer func() {
		Log = originalLogger
	}()

	// Test initialization
	Init()
	if Log == nil {
		t.Error("Expected logger to be initialized")
	}

	// Test logging functionality
	Log.Info("test message")
	Log.Error("test error", zap.Error(ErrInitLogger))
}
