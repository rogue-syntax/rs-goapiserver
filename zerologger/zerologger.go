package zerologger

import (
	"io"
	"os"
	"sync"
	"time"

	zerolog "github.com/rogue-syntax/rs_zerolog"
	"github.com/rogue-syntax/rs_zerolog/pkgerrors"
	"gopkg.in/natefinch/lumberjack.v2"
)

var once sync.Once

var error_logger zerolog.Logger

var debug_logger zerolog.Logger

var info_logger zerolog.Logger

func GetErrorLogger() zerolog.Logger {
	once.Do(func() {
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		zerolog.ErrorFuncMarshaler = pkgerrors.MarshalCallingFunction
		zerolog.TimeFieldFormat = time.RFC3339Nano

		logLevel := int(zerolog.ErrorLevel)

		var output io.Writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}

		if os.Getenv("APP_ENV") != "development" {
			fileLogger := &lumberjack.Logger{
				Filename:   "/var/logs/apiserver.log",
				MaxSize:    5, //
				MaxBackups: 10,
				MaxAge:     14,
				Compress:   true,
			}

			output = zerolog.MultiLevelWriter(os.Stderr, fileLogger)
		}

		error_logger = zerolog.New(output).Level(zerolog.Level(logLevel)).With().Timestamp().Logger()
	})

	return error_logger
}

func LogError(err *error, msg string) {
	error_logger.Error().CallingFunc().Stack().Err(*err).Msg(msg)
}
