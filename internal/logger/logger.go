package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"time"
)

const (
	colorCyan = "\033[36m"
)

func colorTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	coloredTime := fmt.Sprintf("%s%s", colorCyan, t.Format("2006-01-02 15:04:05"))
	enc.AppendString(coloredTime)
}

var logger *zap.Logger

func setUpSugaredLogger() *zap.SugaredLogger {
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:          "time",
		LevelKey:         "level",
		NameKey:          "Logger",
		CallerKey:        "caller",
		MessageKey:       "msg",
		StacktraceKey:    "stacktrace",
		EncodeTime:       colorTimeEncoder,
		EncodeLevel:      zapcore.CapitalColorLevelEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
		ConsoleSeparator: " | ",
	}

	consoleEncoder := zapcore.NewConsoleEncoder(encoderCfg)

	core := zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), zapcore.DebugLevel)

	logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger.Sugar()
}

func Sync() {
	if logger != nil {
		_ = logger.Sync()
	}
}

type MyLogger struct {
	*zap.SugaredLogger
}

func newLogger() *MyLogger {

	logger := setUpSugaredLogger()

	return &MyLogger{
		logger,
	}
}

// имплементируем метод интерфейса BotLogger
func (l *MyLogger) Println(v ...interface{}) {
	l.Info(fmt.Sprint(v...))
}

// имплементируем метод интерфейса BotLogger
func (l *MyLogger) Printf(format string, v ...interface{}) {
	l.Info(fmt.Sprintf(format, v...))
}

var Instance = newLogger()
