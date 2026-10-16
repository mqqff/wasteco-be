package log

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

type LogInfo map[string]interface{}

var logger zerolog.Logger

func init() {
	writer := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	logger = zerolog.New(writer).With().Timestamp().Logger()
}

func GetLogger() *zerolog.Logger {
	return &logger
}

func Info(fields LogInfo, msg string) {
	event := logger.Info()
	for key, value := range fields {
		event = event.Interface(key, value)
	}
	event.Msg(msg)
}

func Error(fields LogInfo, msg string) {
	event := logger.Error()
	for key, value := range fields {
		event = event.Interface(key, value)
	}
	event.Msg(msg)
}
