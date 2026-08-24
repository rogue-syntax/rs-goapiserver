package ev_errors

import (
	"net/http"

	"github.com/rogue-syntax/rs-goapiserver/apireturn/apierrorkeys"
	"github.com/rogue-syntax/rs-goapiserver/rs_ev_src"

	"github.com/rogue-syntax/rs-goapiserver/apierrors"
)

func HandleEVError(evError rs_ev_src.EVEventExecError, r *http.Request) error {
	if evError.ActionError != nil {
		//return the action error to caller
		return evError.ActionError
	}

	//silenetly log the others
	if evError.StreamError != nil {
		apierrors.HandleError(r, evError.StreamError, apierrorkeys.EventError, nil)

	}
	if evError.StoreError != nil {
		apierrors.HandleError(r, evError.StoreError, apierrorkeys.EventError, nil)

	}
	return nil
}
