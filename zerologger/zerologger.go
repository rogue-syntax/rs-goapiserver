package zerologger

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/rogue-syntax/rs-goapiserver/rs_go_requestlogger"
	zerolog "github.com/rogue-syntax/rs_zerolog"
	"github.com/rogue-syntax/rs_zerolog/pkgerrors"
	"gopkg.in/natefinch/lumberjack.v2"
)

var once sync.Once

var error_logger zerolog.Logger

var req_logger zerolog.Logger

type StringWriter struct {
	bBuffer bytes.Buffer
	Pbytes  []byte
	Sstring string
}

func (sw *StringWriter) Write() (n int, err error) {
	return sw.bBuffer.Write(sw.Pbytes)
}

func (sw *StringWriter) WriteString() (n int, err error) {
	return sw.bBuffer.Write(sw.Pbytes)
}

func GetStringWriter() *StringWriter {
	return &StringWriter{}
}

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

		//if os.Getenv("APP_ENV") != "development" {
		fileLogger := &lumberjack.Logger{
			Filename:   "/var/logs/apiserver.log",
			MaxSize:    5, //
			MaxBackups: 10,
			MaxAge:     14,
			Compress:   true,
		}

		output = zerolog.MultiLevelWriter(os.Stderr, fileLogger)

		//}

		error_logger = zerolog.New(output).Level(zerolog.Level(logLevel)).With().Timestamp().Logger()
	})

	return error_logger
}

func LogError(err *error, msg string, r *http.Request) string {
	if r != nil {
		log_id, _ := rs_go_requestlogger.CtxGetReqId(r.Context())
		logStr := error_logger.Error().CallingFunc().Stack().Err(*err).Str("Req_id", log_id).Msg(msg)
		return logStr
	}
	logStr := error_logger.Error().CallingFunc().Stack().Err(*err).Msg(msg)
	return logStr
}

var ReqLogger = &lumberjack.Logger{
	Filename:   "/var/logs/requestLogger.log",
	MaxSize:    5, //
	MaxBackups: 10,
	MaxAge:     14,
	Compress:   true,
}

var Logger *log.Logger

func GetReqLogger() zerolog.Logger {
	Logger = log.New(ReqLogger, "", log.LstdFlags)
	once.Do(func() {

		zerolog.TimeFieldFormat = time.RFC3339Nano

		logLevel := int(zerolog.NoLevel)

		var output io.Writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}

		//if os.Getenv("APP_ENV") != "development" {

		output = zerolog.MultiLevelWriter(os.Stderr, ReqLogger)

		//}

		req_logger = zerolog.New(output).Level(zerolog.Level(logLevel)).With().Timestamp().Logger()
	})

	return req_logger
}

func LogRequest(msg string) string {
	//log := req_logger.Log().Str("req", msg).Msg(msg)
	msgStr := msg
	ReqLogger.Write([]byte(msgStr))
	return msg
}
