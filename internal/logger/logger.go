package logger

import (
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var Logger zerolog.Logger

// InitLogger 初始化日志
func InitLogger(level, format, output string) error {
	// 设置日志级别
	var logLevel zerolog.Level
	switch level {
	case "debug":
		logLevel = zerolog.DebugLevel
	case "info":
		logLevel = zerolog.InfoLevel
	case "warn":
		logLevel = zerolog.WarnLevel
	case "error":
		logLevel = zerolog.ErrorLevel
	default:
		logLevel = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(logLevel)

	// 设置输出
	var writer io.Writer
	if output == "stdout" || output == "" {
		writer = os.Stdout
	} else {
		file, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		writer = file
	}

	// 设置格式
	if format == "text" {
		writer = zerolog.ConsoleWriter{Out: writer}
	}

	Logger = zerolog.New(writer).With().Timestamp().Logger()
	log.Logger = Logger

	return nil
}

// WithFields 创建带字段的日志
func WithFields(fields map[string]interface{}) *zerolog.Event {
	event := Logger.Info()
	for k, v := range fields {
		switch val := v.(type) {
		case string:
			event = event.Str(k, val)
		case int:
			event = event.Int(k, val)
		case int64:
			event = event.Int64(k, val)
		case bool:
			event = event.Bool(k, val)
		default:
			event = event.Interface(k, val)
		}
	}
	return event
}
