package rs_zxcvbn

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/rogue-syntax/rs-goapiserver/apireturn/apierrorkeys"

	"github.com/nbutton23/zxcvbn-go"
	"github.com/rogue-syntax/rs-goapiserver/apierrors"
	"github.com/rogue-syntax/rs-goapiserver/apireturn"
)

const (
	MIN_PW_STRENGTH_SCORE = 3
)

// PWRecRequest is the request for the password strength check
// @param PwStr is the password string
// @param Uid is the user id
// @param UserInputs is a list of strings that are used to help the password strength check
type PWRecRequest struct {
	PwStr      string
	Uid        int
	UserInputs []string
}

/*
type Returner struct {
	score scoring.MinEntropyMatch
	req   PWRecRequest
}
*/

// @API: /v1/pw-score
//  - @Method: POST
//  - @Desc: Check the strength of a password
//  - @Input: PWRecRequest
//  - @Output: scoring.MinEntropyMatch
//  - @OutputWrapper: JsonReturn

func Handler_RSZxcvbn(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	var pwReq PWRecRequest
	err := json.NewDecoder(r.Body).Decode(&pwReq)
	if err != nil {
		apierrors.HandleError(nil, err, err.Error(), &apierrors.ReturnError{Msg: err.Error(), W: &w})
		return
	}

	score := zxcvbn.PasswordStrength(pwReq.PwStr, pwReq.UserInputs)
	//clear out the password on return
	score.Password = ""
	//ret := Returner{score, pwReq}
	apireturn.ApiJSONReturn(score, apierrorkeys.NOError, &w)
}
