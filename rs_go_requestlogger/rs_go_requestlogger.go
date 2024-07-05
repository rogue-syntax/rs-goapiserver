package rs_go_requestlogger

import (
	"context"
	"net/http"
	"net/url"

	"github.com/pkg/errors"

	"github.com/rogue-syntax/rs-goapiserver/apireturn/apierrorkeys"
)

type RequestVars struct {
	RequestURI string
	Method     string
	RemoteAddr string
	UserAgent  string
	Referer    string
	Proto      string
	Host       string
	Header     http.Header
	PostForm   url.Values
	Cookies    map[string]string
	URL        string
	Body       string
}

type RSRequestLogger struct {
	Endpoint    string
	RequestVars RequestVars
	ErrorLogs   []string
	Req_id      string
}

type keyType string

const idKey keyType = "req_id"

func CtxWithReqId(ctx context.Context, req_id string) context.Context {
	return context.WithValue(ctx, idKey, req_id)
}
func CtxGetReqId(ctx context.Context) (string, error) {
	req_id, ok := ctx.Value(idKey).(string)
	if !ok {
		err := errors.New(apierrorkeys.ContextError)
		return "", err
	}
	return req_id, nil
}

const loggerKey keyType = "RSLogger"

func CtxWithRSLogger(ctx context.Context, rsLogger *RSRequestLogger) context.Context {
	return context.WithValue(ctx, loggerKey, rsLogger)
}
func CtxGetRSLogger(ctx context.Context) (*RSRequestLogger, error) {
	user, ok := ctx.Value(loggerKey).(*RSRequestLogger)
	if !ok {
		err := errors.New(apierrorkeys.ContextError)
		return nil, err
	}
	return user, nil
}
