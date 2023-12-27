package apierrors

import (
	"net/http"

	"github.com/rogue-syntax/rs-goapiserver/apireturn"
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

apierrors.HandleError(err, "ikyc.Handler_kycInitiate", apiErrorKeys.JSONMarshalError, "3", w, r)

Takes reqError ReqError, w *http.ResponseWriter, r *http.Request (w & r are nil if not handling an API route)

reqError:ReqError is:

  - ErrObj: The error Object being caught
  - PackageName: Name of the calling package
  - FuncName: Name of the calling function
  - ErrorKey:  A const string from apiErrorKeys package
  - UniqueId: A unique id to help identify error origination in source code
  - RetError: An emun to convet error code to calling agent, 0 is no error
*/
/*func HandleError(reqError ReqError, w *http.ResponseWriter, r *http.Request) {
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
	Data *interface{}
	W    *http.ResponseWriter
}

/*
HandleError
Logs the error and its message to logger system, gives the ReturnError to ApiJSONReturn to give back to the client / frontend
-	ReturnError.Data should be nil or an object ref to serialize
-	ReturnError.Msg should be a const error code i.e. AUTHORIZATION_ERROR to inform client of necessary response to take
*/
func HandleError(err error, msg string, returnError *ReturnError) {
	//log the error to logger, give the return error to ApiJSONReturn to give back to the client / frontend
	zerologger.LogError(&err, msg)
	if returnError != nil {
		//Data should be nil or an object ref to serialize
		//
		apireturn.ApiJSONReturn(returnError.Data, returnError.Msg, returnError.W)
	}
}
