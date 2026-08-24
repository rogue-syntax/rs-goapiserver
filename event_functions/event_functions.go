package event_functions

import (
	"net/http"

	"github.com/pkg/errors"
	"github.com/rogue-syntax/rs-goapiserver/ev_errors"
	"github.com/rogue-syntax/rs-goapiserver/rs_ev_src"
	"github.com/rogue-syntax/rs-goapiserver/rs_go_requestlogger"
)

func CreateAndCallEVEvent[DATATYPE any, CONTEXTTYPE any, RETURNTYPE any](r *http.Request, co_id int, action rs_ev_src.IEVAction[DATATYPE, RETURNTYPE], data *DATATYPE, metaData *CONTEXTTYPE, eventConst rs_ev_src.EVTypes_int, scheduleType rs_ev_src.EVScheduleTypes_int, returnConst RETURNTYPE) (RETURNTYPE, error) {
	req_id, _ := rs_go_requestlogger.CtxGetReqId(r.Context())
	var ev rs_ev_src.EVEvent[DATATYPE, CONTEXTTYPE, RETURNTYPE]
	ev.SetEVEvent(
		co_id,
		action,
		data,
		metaData,
		eventConst,
		scheduleType,
		req_id,
	)
	_, evError := rs_ev_src.DoEVEventAction[DATATYPE](&ev)
	err := ev_errors.HandleEVError(evError, r)
	if err != nil {
		return returnConst, errors.WithStack(err)
	}
	return returnConst, nil
}
