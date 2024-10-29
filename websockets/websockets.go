package websockets

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/rogue-syntax/rs-goapiserver/apicontext"
	"github.com/rogue-syntax/rs-goapiserver/apierrors"
	"github.com/rogue-syntax/rs-goapiserver/apireturn/apierrorkeys"

	"github.com/gorilla/websocket"
)

type GenericEvents struct {
	EventKey  int
	EventName string
}

const (
	CONNECTED = iota
	DISCONNECTED
	TEST
)

var GenericEventsMap = map[int]SocketEvent{
	CONNECTED:    {EventKey: CONNECTED, EventName: "CONNECTED", Data: nil},
	DISCONNECTED: {EventKey: DISCONNECTED, EventName: "DISCONNECTED", Data: nil},
	TEST:         {EventKey: TEST, EventName: "TEST", Data: nil},
}

type SocketEvent struct {
	EventKey  int
	EventName string
	Data      interface{}
}

type IncomingSocketEvent struct {
	User_id     int
	SocketEvent SocketEvent
	Conn_id     uuid.UUID
}

type ProgressEvent struct {
	EventName string
	Position  int
	Start     int
	End       int
}

// return a simple message to the user who calls this endpoint
func TestWS(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	usr, err := apicontext.CtxGetUser(ctx)
	if err != nil {
		apierrors.HandleError(nil, err, err.Error(), &apierrors.ReturnError{Msg: apierrorkeys.AuthorizationError, W: &w})
		return
	}
	msg := GenericEventsMap[TEST]
	Channel_WriteToUserSockets([]int{usr.User_id}, &msg)

}

// call this endpoint to establish a websocket connection
func WsEndpoint(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	usr, err := apicontext.CtxGetUser(ctx)
	if err != nil {
		apierrors.HandleError(nil, err, err.Error(), &apierrors.ReturnError{Msg: apierrorkeys.AuthorizationError, W: &w})
		return
	}

	req := WebSocketChanReq{Type: AddConn, User_ids: []int{(*usr).User_id}, Response: make(chan WebSocketChanResp), W: &w, R: r}
	wsChannel <- req
	resp := <-req.Response
	if resp.Err != nil {
		apierrors.HandleError(nil, resp.Err, apierrorkeys.WebSocketError, &apierrors.ReturnError{Msg: apierrorkeys.WebSocketError, W: &w})
		return
	}
}

func Reader(conn *websocket.Conn, r *http.Request) {
	for {
		// read in a message
		messageType, p, err := conn.ReadMessage()
		if err != nil {

			return
		}
		_ = messageType
		// print out that message for clarity
		fmt.Println(string(p))
		cookie, _ := r.Cookie("session")

		sessionBy := []byte(string(p) + " - " + cookie.Value)
		if err := conn.WriteMessage(1, sessionBy); err != nil {
			return
		}

	}
}

// send a message user(s)
// []int{1,2,3} will send the message to users with user_id 1, 2, and 3
func Channel_WriteToUserSockets(user_ids []int, msg *SocketEvent) error {
	msgBytes, _ := json.Marshal(msg)
	msgStr := string(msgBytes)
	req := WebSocketChanReq{Type: SendMsg, User_ids: user_ids, Msg: msgStr, Response: make(chan WebSocketChanResp)}
	wsChannel <- req
	resp := <-req.Response
	if resp.Err != nil {
		return resp.Err
	}
	return nil
}

func Channel_GetUserSockets() (map[int][]*RSSocketConnection, error) {
	var returnMap map[int][]*RSSocketConnection
	req := WebSocketChanReq{Type: GetAllSockets, User_ids: nil, Msg: "", Response: make(chan WebSocketChanResp)}
	wsChannel <- req
	resp := <-req.Response
	if resp.Err != nil {

		return returnMap, resp.Err
	}
	returnMap = resp.Msg.(map[int][]*RSSocketConnection)
	return returnMap, nil
}

func ExampleOfProgressUpdate(userSock *websocket.Conn) {
	var progEV ProgressEvent
	progEV.EventName = "propSearchProgEv"
	progEV.Start = 0
	progEV.End = 100

	if userSock != nil {
		progEvBytes, _ := json.Marshal(progEV)
		userSock.WriteMessage(1, progEvBytes)
	}
}
