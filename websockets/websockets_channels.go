package websockets

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

func IntiWebsockets(ch chan WebSocketChanReq, incoming chan IncomingSocketEvent, logProvider LogEventProvider, disconnectCallback func(*RSSocketConnection)) {
	Disconnector.SetDisconnectCallback(disconnectCallback)
	SetChannels(ch, incoming)
	SetLogEventProvider(logProvider)
	go ManageWebSockets(ch, incoming)
	go HandleIncomingWebSockets(incoming)

}

type WebsocketDisconnecter struct {
	DisconnectCallback func(*RSSocketConnection)
}

func (wd *WebsocketDisconnecter) SetDisconnectCallback(callback func(*RSSocketConnection)) {
	wd.DisconnectCallback = callback
}

var Disconnector WebsocketDisconnecter

type WSCallbackFunc func(*IncomingSocketEvent)

var WSCallbacks = make(map[string]WSCallbackFunc)

func RegisterCallback(eventName string, callback WSCallbackFunc) {
	WSCallbacks[eventName] = callback
}

type LogEventProvider interface {
	LogMessage(err error, msg string, data interface{}) int
}

var logEventProvider LogEventProvider

func SetLogEventProvider(provider LogEventProvider) {
	logEventProvider = provider
}

type WSChannelErrorKeys string

const (
	WebSocketError       WSChannelErrorKeys = "WebSocket Error"
	GoRoutineRecovery    WSChannelErrorKeys = "GoRoutine Recovery"
	SocketEventUnmarshal WSChannelErrorKeys = "Socket Event Unmarshal Error"
)

// the UserSocket channel to manage the websockets
var wsChannel chan WebSocketChanReq

var incomingMessageChannel chan IncomingSocketEvent

// UserSocketType is a map of user_id to their connections
type UserSocketType map[int][]*RSSocketConnection

// upgrader for the websocket connection
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type BasicSocketEvents string

const (
	PING BasicSocketEvents = "PING"
	PONG BasicSocketEvents = "PONG"
)

var BasicSocketEventMap = map[BasicSocketEvents]SocketEvent{
	PING: {EventKey: 0, EventName: "PING", Data: nil},
	PONG: {EventKey: 1, EventName: "PONG", Data: nil},
}

//need a map of callback functions for incoming websocket messages
//map of event name to function
//but the function should be able to reside within other packages
//...

var SocketEventMap = map[string]func(*websocket.Conn, *SocketEvent) error{
	"CONNECTED": func(ws *websocket.Conn, se *SocketEvent) error {
		return nil
	},
}

// WebSocketChanRespType is a type for the response type of the websocket channel
type WSChannelRespType int

// keys for interpreting the response data. is it a simple message, or a SomeObjectKind struct?
const (
	ErrorKey WSChannelRespType = iota // simple error response type from apierrorkeys
	SuccessMsg
)

type WSChannelReqType int

const (
	AddConn WSChannelReqType = iota
	SendMsg
	GetAllSockets
	UpdateUrl
)

type WebSocketChanResp struct {
	Err  error
	Msg  any
	Type WSChannelRespType //to infer the type of response
}

// SetChannel allows other packages to set the wsChannel
func SetChannels(ch chan WebSocketChanReq, incoming chan IncomingSocketEvent) {
	wsChannel = ch
	incomingMessageChannel = incoming
}

// Operation request struct
type WebSocketChanReq struct {
	Type     WSChannelReqType
	User_ids []int
	Msg      string
	Response chan WebSocketChanResp
	W        *http.ResponseWriter
	R        *http.Request
	Sockets  chan map[int][]RSSocketConnection
}

type RSSocketConnection struct {
	Conn    *websocket.Conn
	User_id int
	Conn_id uuid.UUID
}

// The managing goroutine function
// This function manages the websockets, and is called by the main server
// It listens for requests on the wsChannel, and acts on them
// The requests are of type WebSocketChanReq
// The function manages the connections of the users
func ManageWebSockets(requestChannel chan WebSocketChanReq, incomingMsgChannel chan IncomingSocketEvent) {
	//map of user to their connections

	//UserSockets := make(map[int][]*websocket.Conn)
	UserSockets := make(map[int][]*RSSocketConnection)

	userSocketsMutex := &sync.Mutex{}

	//weak check origin for now, endpoint is guarded by authentication
	//same origin or sother check can be implemented to return true
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from crash in ManageWebSockets:", r)
			rString := fmt.Sprint(r)
			logEventProvider.LogMessage(
				errors.Wrap(errors.New(rString),
					"websockets.ManageWebsockets recovered with data: "+rString),
				string(GoRoutineRecovery),
				nil)
			// Decide to restart the goroutine
			go ManageWebSockets(requestChannel, incomingMsgChannel)
		}
	}()

	for req := range requestChannel {
		switch req.Type {
		case AddConn:
			conn, resp := RegisterConn(UserSockets, req.R, *req.W, req.User_ids[0], requestChannel)
			if conn != nil {
				go listenOnWebSocket(conn, req.User_ids[0], incomingMsgChannel, UserSockets, userSocketsMutex)
			}
			req.Response <- resp
			close(req.Response)
		case SendMsg:
			resp := MsgUsers(UserSockets, req.User_ids, &req.Msg)
			req.Response <- resp
		case GetAllSockets:
			resp := WebSocketChanResp{Err: nil, Msg: UserSockets, Type: SuccessMsg}
			req.Response <- resp
			close(req.Response)

		}

	}
}

func listenOnWebSocket(conn *RSSocketConnection, usr_id int, incomingMsgChannel chan IncomingSocketEvent, UserSockets UserSocketType, userSocketsMutex *sync.Mutex) {
	defer conn.Conn.Close()
	for {
		_, msg, err := conn.Conn.ReadMessage()
		if err != nil {
			break
			//clean up the connection with write to user sockets which closes any it cant write to
		}
		var socketEvent SocketEvent
		err = json.Unmarshal(msg, &socketEvent)
		if err != nil {
			logEventProvider.LogMessage(err, string(SocketEventUnmarshal), nil)
			continue
		}
		fmt.Println("Incoming Messgage")
		if cbFunc, ok := WSCallbacks[socketEvent.EventName]; ok {
			fmt.Println("Handling Incoming Message")
			cbFunc(&IncomingSocketEvent{User_id: usr_id, SocketEvent: socketEvent, Conn_id: conn.Conn_id})
		}
		//incomingMsgChannel <- IncomingSocketEvent{User_id: usr_id, SocketEvent: socketEvent}
	}

	fmt.Println("Client Disconnected")
	msg := GenericEventsMap[DISCONNECTED]
	msgBytes, _ := json.Marshal(&msg)
	msgStr := string(msgBytes)
	userSocketsMutex.Lock()
	WriteToUserSockets(UserSockets, usr_id, &msgStr)
	userSocketsMutex.Unlock()
	Disconnector.DisconnectCallback(conn)
}
func RegisterConn(UserSockets UserSocketType, r *http.Request, w http.ResponseWriter, usr_id int, reqs chan WebSocketChanReq) (*RSSocketConnection, WebSocketChanResp) {
	var wsu RSSocketConnection
	var socketList []*RSSocketConnection
	if sockets, ok := UserSockets[usr_id]; !ok {
		UserSockets[usr_id] = socketList
	} else {
		socketList = sockets
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	wsu.Conn = ws
	if err != nil {
		resp := WebSocketChanResp{Err: err, Msg: WebSocketError, Type: ErrorKey}
		return &wsu, resp
	}
	wsu.Conn_id = uuid.New()
	wsu.User_id = usr_id
	UserSockets[usr_id] = append(socketList, &wsu)
	msgBytes, _ := json.Marshal(SocketEvent{EventKey: GenericEventsMap[CONNECTED].EventKey, EventName: GenericEventsMap[CONNECTED].EventName, Data: "Client Connected"})
	msg := string(msgBytes)
	err = WriteToUserSockets(UserSockets, usr_id, &msg)
	if err != nil {
		resp := WebSocketChanResp{Err: err, Msg: WebSocketError, Type: ErrorKey}
		return &wsu, resp
	}
	resp := WebSocketChanResp{Err: nil, Msg: "success", Type: SuccessMsg}
	return &wsu, resp
}

func WriteToUserSockets(UserSockets UserSocketType, user_id int, msg *string) error {
	var socketsToRemove []int
	if socketSlice, ok := UserSockets[user_id]; ok {
		for i, ws := range socketSlice {
			err := ws.Conn.WriteMessage(websocket.TextMessage, []byte(*msg))
			if err != nil {
				fmt.Println(err)
				closeErr := ws.Conn.Close()
				if closeErr != nil {
					fmt.Println("Error closing WebSocket:", closeErr)
					logEventProvider.LogMessage(closeErr, string(WebSocketError), nil)
				}
				socketsToRemove = append(socketsToRemove, i)
				continue
			}
		}
	}

	for i, idx := range socketsToRemove {
		adjustedIndex := idx - i
		UserSockets[user_id] = append(UserSockets[user_id][:adjustedIndex], UserSockets[user_id][adjustedIndex+1:]...)
	}
	return nil
}

func MsgUsers(UserSockets UserSocketType, user_ids []int, msg *string) WebSocketChanResp {
	for _, user_id := range user_ids {
		err := WriteToUserSockets(UserSockets, user_id, msg)
		if err != nil {
			resp := WebSocketChanResp{Err: err, Msg: WebSocketError, Type: ErrorKey}
			return resp
		}
	}
	resp := WebSocketChanResp{Err: nil, Msg: "success", Type: SuccessMsg}
	return resp
}

func HandleIncomingWebSockets(incomingMsgChannel chan IncomingSocketEvent) {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from crash in HandleIncomingWebSockets:", r)
			rString := fmt.Sprint(r)
			logEventProvider.LogMessage(
				errors.Wrap(errors.New(rString),
					"websockets.HandleIncomingWebSockets recovered with data: "+rString),
				string(GoRoutineRecovery),
				nil)
			// Decide to restart the goroutine
			go HandleIncomingWebSockets(incomingMsgChannel)
		}
	}()

	for {
		select {
		case msg := <-incomingMsgChannel:

			if cbFunc, ok := WSCallbacks[msg.SocketEvent.EventName]; ok {
				fmt.Println("Handling Incoming Message")
				cbFunc(&msg)
			} else {
				logEventProvider.LogMessage(
					errors.New("websockets.HandleIncomingWebSockets received an event with no callback function: "+msg.SocketEvent.EventName),
					string(WebSocketError),
					nil)

			}
		}
	}
}
