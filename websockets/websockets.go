package websockets

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"rs-apiserver.com/apicontext"
	"rs-apiserver.com/apierrors"
	"rs-apiserver.com/apireturn/apierrorkeys"
	"rs-apiserver.com/authentication"
	"rs-apiserver.com/entities/user"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type SocketEvent struct {
	EventKey string
	Data     interface{}
}

var UserSockets map[int]map[string]*websocket.Conn

func InitWebSockets() {

	UserSockets = make(map[int]map[string]*websocket.Conn)
}

type ProgressEvent struct {
	EventName string
	Position  int
	Start     int
	End       int
}

func TestWS(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	usr, err := apicontext.CtxGetUser(ctx)
	if err != nil {
		apierrors.HandleError(err, err.Error(), &apierrors.ReturnError{Msg: apierrorkeys.AuthorizationError, W: &w})
		return
	}
	fmt.Fprintf(w, "big tings")

	usb, _ := json.Marshal(UserSockets)

	fmt.Fprintf(w, string(usb))

	user_agent := authentication.GetUserAgentHashFromRequest(r)

	if ws, ok := UserSockets[*&usr.User_id][user_agent]; ok {
		//if ws, ok := userSockets[(*usr).UID]; ok {
		time.Sleep(5 * time.Second)

		err := ws.WriteMessage(1, []byte(`{ "action":"ethPayTxCB", "msg":"success"}`))
		if err != nil {
			fmt.Fprintf(w, err.Error())
		}

		time.Sleep(8 * time.Second)

		err = ws.WriteMessage(1, []byte(`{ "action":"payTokenTxCB", "msg":"success"}`))
		if err != nil {
			fmt.Fprintf(w, err.Error())
		}

	}
}

func GetUserSocket(usr *user.UserExternal, r *http.Request) *websocket.Conn {
	user_agent := authentication.GetUserAgentHashFromRequest(r)
	if wsu, ok := UserSockets[(*usr).User_id][user_agent]; ok {
		return wsu
	}
	return nil
}

func WsEndpoint(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	usr, err := apicontext.CtxGetUser(ctx)
	if err != nil {
		apierrors.HandleError(err, err.Error(), &apierrors.ReturnError{Msg: apierrorkeys.AuthorizationError, W: &w})
		return
	}
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	// upgrade this connection to a WebSocket
	// connection
	user_agent := authentication.GetUserAgentHashFromRequest(r)
	//socket exists
	/*_, isUser :=  userSockets[(*usr).User_id]
	if !isUser {
		userSockets[(*usr).User_id]
	}*/
	if wsu, ok := UserSockets[(*usr).User_id][user_agent]; ok {
		//close existing socket, replace with new one
		wsu.Close()
		wsu, err := upgrader.Upgrade(w, r, nil)
		if err != nil {

		}

		UserSockets[(*usr).User_id][user_agent] = wsu

		msg, _ := json.Marshal(SocketEvent{EventKey: "CONNECTION", Data: "Client Connected"})
		_ = wsu.WriteMessage(1, msg)
		Reader(wsu, r)

	} else {
		//socket doesnt exist
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {

		}
		UserSockets[(*usr).User_id] = make(map[string]*websocket.Conn)
		UserSockets[(*usr).User_id][user_agent] = ws
		msg, _ := json.Marshal(SocketEvent{EventKey: "CONNECTION", Data: "Client Connected"})
		_ = ws.WriteMessage(1, msg)
		Reader(ws, r)
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
