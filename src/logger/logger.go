package logger

import (
	"fmt"
	"log/slog"
)

var Logger = newLogger()

type SlogLogger struct {
	logger *slog.Logger
}

func newLogger() *SlogLogger {

	logger := SetupPrettySlog()

	return &SlogLogger{
		logger: logger,
	}
}

// имплементируем метод интерфейса BotLogger
func (l *SlogLogger) Println(v ...interface{}) {
	l.logger.Info(fmt.Sprint(v...))
}

// имплементируем метод интерфейса BotLogger
func (l *SlogLogger) Printf(format string, v ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, v...))
}

func Error(msg string, args ...any) {
	Logger.logger.Error(msg, args...)
}

func Info(msg string, args ...any) {
	Logger.logger.Info(msg, args...)
}

func Debug(msg string, args ...any) {
	Logger.logger.Debug(msg, args...)
}

func Warn(msg string, args ...any) {
	Logger.logger.Warn(msg, args...)
}
