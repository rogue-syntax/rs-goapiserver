package apireturn

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rogue-syntax/rs-goapiserver/apireturn/apierrorkeys"
	"github.com/rogue-syntax/rs-goapiserver/entities/user"
	"github.com/rogue-syntax/rs-goapiserver/zerologger"
)

const (
	USER_NOT_AUTHENTICATED   = iota
	SEND_MAIL_ERROR          = iota
	BAD_REQUEST_PAYLOAD      = iota
	SYSTEM_ENCOUNTERED_ERROR = iota
	ISSUE_TOKEN_ERROR        = iota
)

type ApiParamaterDefinition struct {
	ParamKey        string
	ParamDefinition string
}

type ApiDefinition struct {
}

type JsonReturn struct {
	Error string
	Data  interface{}
}

type JsonResponse struct {
	Error string
	Data  json.RawMessage
}

type AuthenicatedJsonReturn struct {
	Error string
	User  *user.UserExternal
	Data  interface{}
}

/*
API JSON Return
  - jsonRet : Data to be returned. Can be Interface object, string, int, or nil
  - errMsg : Const string error code from apireturn.AUTH_ERROR
*/

func ApiJSONReturn(jsonRet interface{}, errMsg string, w *http.ResponseWriter) {
	if w != nil {
		jr := JsonReturn{Error: errMsg, Data: jsonRet}
		jrStr, err := json.Marshal(jr)
		if err != nil {
			zerologger.LogError(&err, apierrorkeys.JSONMarshalError)
			fmt.Fprint(*w, `{ "Error":"`+apierrorkeys.JSONMarshalError+`", Data:"`+err.Error()+`"}`)
		}
		fmt.Fprint(*w, string(jrStr))
	}
}
