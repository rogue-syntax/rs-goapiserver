package apierrors

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
	"github.com/rogue-syntax/rs-goapiserver/apireturn"
	"github.com/rogue-syntax/rs-goapiserver/apireturn/apierrorkeys"
	"github.com/rogue-syntax/rs-goapiserver/rs_go_requestlogger"
	"github.com/rogue-syntax/rs-goapiserver/zerologger"
)

type ErrorWrapper struct {
	UnixTime int64
	Time     string
	Key      string
	Info     ErrorInfo
}

type ErrorInfo struct {
	Msg          string
	PackageName  string
	FunctionName string
	UniqueId     string
	ErrorObj     error
	//stackError := errors.Wrap(errorObj, 1)
}

type ReqError struct {
	ErrorObj    error
	PackageName string
	FuncName    string
	ErrorKey    string
	UniqueId    string
	RetError    int
}

/*
a general purpose error handler, logging erros to std out and also returning message to sender

Example:

apierrors.HandleError(nil, err, "ikyc.Handler_kycInitiate", apiErrorKeys.JSONMarshalError, "3", w, r)

Takes reqError ReqError, w *http.ResponseWriter, r *http.Request (w & r are nil if not handling an API route)

reqError:ReqError is:

  - ErrObj: The error Object being caught
  - PackageName: Name of the calling package
  - FuncName: Name of the calling function
  - ErrorKey:  A const string from apiErrorKeys package
  - UniqueId: A unique id to help identify error origination in source code
  - RetError: An emun to convet error code to calling agent, 0 is no error
*/
/*func HandleError(nil, reqError ReqError, w *http.ResponseWriter, r *http.Request) {
	errorInfo := ErrorInfo{
		Msg:          reqError.ErrorObj.Error(),
		PackageName:  reqError.PackageName,
		FunctionName: reqError.FuncName,
		UniqueId:     reqError.UniqueId,
		ErrorObj:     reqError.ErrorObj,
	}
	errWrapper := ErrorWrapper{
		UnixTime: time.Now().Unix(),
		Time:     time.Now().String(),
		Key:      reqError.ErrorKey,
		Info:     errorInfo,
	}
	apiErorJson, _ := json.Marshal(errWrapper)
	fmt.Printf(string(apiErorJson) + ",\n")
	if w != nil {
		apireturn.ApiJSONReturn("", apierrorkeys.APIReqError, w)
	}
}*/

type ReturnError struct {
	Msg  string
	Data interface{}
	W    *http.ResponseWriter
}

type LogError struct {
	ErrType string
	Data    interface{}
}

func NewLogError(errType string, data interface{}) string {
	logErr := LogError{
		ErrType: errType,
		Data:    data,
	}
	jb, err := json.Marshal(&logErr)
	if err != nil {
		HandleError(nil, err, apierrorkeys.LogGenError, nil)
	}
	return string(jb)
}

type DataSliceInterface interface {
	Interf() int
}

type DataSliceItem struct {
	Data interface{}
}

func (dsi DataSliceItem) Interf() int {
	return 0
}

type JsonArrayError struct {
	DataSlice []interface{}
}
type JsonMapError struct {
	DataMap map[string]interface{}
}

func (jsonError JsonArrayError) Error() string {
	jb, err := json.Marshal(jsonError.DataSlice)
	if err != nil {
		return `{"logError":"` + err.Error() + `"}`
	}
	return string(jb)
}

func (jsonError JsonMapError) Error() string {
	jb, err := json.Marshal(jsonError.DataMap)
	if err != nil {
		return `{"logError":"` + err.Error() + `"}`
	}
	//str, _ := strconv.Unquote(string(jb))
	return string(string(jb))
}

/*
var jsonError apierrors.JsonError
		jsonError.DataSlice = append(jsonError.DataSlice, err)
		jsonError.DataSlice = append(jsonError.DataSlice, usr)
		jsonError.DataSlice = append(jsonError.DataSlice, userAccount)
*/

func LogJsonArray(args ...any) JsonArrayError {
	var jsonError JsonArrayError
	jsonError.DataSlice = append(jsonError.DataSlice, args...)
	return jsonError
}

func LogJsonMap(mapped map[string]interface{}) JsonMapError {
	var jsonError JsonMapError
	errors.WithStack(jsonError)
	jsonError.DataMap = mapped
	return jsonError
}

type ReqLogStreamCallback func(rsLog *rs_go_requestlogger.RSRequestLogger, logStr string) string
type ReqLogWriteCallback func(rsLog *rs_go_requestlogger.RSRequestLogger) string

type RequestLogStreamer interface {
	Stream(rsLog *rs_go_requestlogger.RSRequestLogger, logStr string) string
	Write(rsLog *rs_go_requestlogger.RSRequestLogger) string
}

// ErrorHandler
type RequestLogHandler struct {
}

func (rlh *RequestLogHandler) Stream(rsLog *rs_go_requestlogger.RSRequestLogger, logStr string) string {
	return ""
}

func (rlh *RequestLogHandler) Write(rsLog *rs_go_requestlogger.RSRequestLogger) string {
	jb, _ := json.Marshal(&rsLog)
	msgStr := string(jb)
	return zerologger.LogRequest(msgStr)
}

type ErrorLogStreamCallback func(err error, msg string, jsonError string, r *http.Request) string
type ErrorLogWriteCallback func(err error, msg string, r *http.Request) string

type ErrorLogStreamer interface {
	Stream(err error, msg string, jsonError string, r *http.Request) string
	Write(err error, msg string, r *http.Request) string
}

// ErrorHandler
type ErrorLogHandler struct {
}

func (rlh *ErrorLogHandler) Stream(err error, msg string, jsonError string, r *http.Request) string {
	return ""
}

func (rlh *ErrorLogHandler) Write(err error, msg string, r *http.Request) string {
	err = errors.WithStack(err)
	return zerologger.LogError(&err, msg, nil)
}

type RequestLogImpl struct {
	ReqHandlerImpl RequestLogStreamer
}

type ErrorLogImpl struct {
	ErrorHandlerImpl ErrorLogStreamer
}

// update RequestLogCallbacks with a custom ErrorLogStreamer interface:
//   - Stream(err error, msg string, jsonError string) string
//   - Write(err error, msg string) string
var RequestLogCallbacks RequestLogImpl

// update ErrorLogCallbacks with a custom ErrorLogStreamer interface:
//   - Stream(err error, msg string, jsonError string) string
//   - Write(err error, msg string) string
var ErrorLogCallbacks ErrorLogImpl

// InitAPIErrorHandlers
// Initialize the APIErrorHandlers with custom RequestLogHandler and ErrorLogHandler
// If customReqLogHandler is nil, a default RequestLogHandler will be used
// If customErrorLogHandler is nil, a default ErrorLogHandler will be used
func InitAPIErrorHandlers(customReqLogHandler RequestLogStreamer, customErrorLogHandler ErrorLogStreamer) {
	if customReqLogHandler != nil {
		RequestLogCallbacks.ReqHandlerImpl = customReqLogHandler
	} else {
		var requestLogHandler RequestLogHandler
		RequestLogCallbacks.ReqHandlerImpl = &requestLogHandler
	}
	if customErrorLogHandler != nil {
		ErrorLogCallbacks.ErrorHandlerImpl = customErrorLogHandler
	} else {
		var errorLogHandler ErrorLogHandler
		ErrorLogCallbacks.ErrorHandlerImpl = &errorLogHandler
	}
}

/*
HandleError
Logs the error and its message to logger system, gives the ReturnError to ApiJSONReturn to give back to the client / frontend
-	ReturnError.Data should be nil or an object ref to serialize
-	ReturnError.Msg should be a const error code i.e. AUTHORIZATION_ERROR to inform client of necessary response to take
*/
func HandleError(r *http.Request, err error, msg string, returnError *ReturnError) {

	errorStr := ErrorLogCallbacks.ErrorHandlerImpl.Write(err, msg, r)
	ErrorLogCallbacks.ErrorHandlerImpl.Stream(err, msg, errorStr, r)

	if r != nil {
		rsLog, _ := rs_go_requestlogger.CtxGetRSLogger(r.Context())
		rsLog.ErrorLogs = append(rsLog.ErrorLogs, err.Error())
		ctx := rs_go_requestlogger.CtxWithRSLogger(r.Context(), rsLog)
		r.WithContext(ctx)
	}

	if returnError != nil {
		apireturn.ApiJSONReturn(returnError.Data, returnError.Msg, returnError.W)
	}
}

func HandleReqLog(r *http.Request) {
	if r != nil {
		rsLog, _ := rs_go_requestlogger.CtxGetRSLogger(r.Context())
		logStr := RequestLogCallbacks.ReqHandlerImpl.Write(rsLog)
		RequestLogCallbacks.ReqHandlerImpl.Stream(rsLog, logStr)
	}

}

//GLOBAL ERROR TYPES

type QueryError struct {
	User_id     int
	Co_id       int
	QueryString string
	QueryError  string
	Values      []interface{}
}
