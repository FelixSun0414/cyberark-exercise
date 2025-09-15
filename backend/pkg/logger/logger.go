package logger

import (
	"log/slog"
	"os"
)

var loggerSingleton *slog.Logger

func init() {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	loggerSingleton = slog.New(handler)
}

func Info(msg string, args ...any) {
	loggerSingleton.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	loggerSingleton.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	loggerSingleton.Error(msg, args...)
}

func Fatal(msg string, args ...any) {
	Error(msg, args...)
	os.Exit(1)
}

func Panic(msg string, args ...any) {
	Error(msg, args...)
	panic(msg)
}
