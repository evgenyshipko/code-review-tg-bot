package logger

import (
	"fmt"
	"log/slog"
	"runtime"
	"strings"
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

func GetStackTraceAsSlice() []string {
	buf := make([]byte, 1024)
	for {
		n := runtime.Stack(buf, false)
		if n < len(buf) {
			// Разбиваем трассировку стека на строки
			return strings.Split(string(buf[:n]), "\n")
		}
		buf = make([]byte, len(buf)*2)
	}
}
