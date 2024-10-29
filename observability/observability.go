package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"

	"github.com/rogue-syntax/rs-goapiserver/websockets"
)

func Handler_LogGoroutineCount(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	fmt.Fprintf(w, "Current number of goroutines: %d", runtime.NumGoroutine())
}

type WebSocketInfo struct {
	User_id     int
	RemoteAddr  string
	Subprotocol string
	NetworkName string
	RSSocket    *websockets.RSSocketConnection
}

func Handler_GetUserSockets(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	userSockets, err := websockets.Channel_GetUserSockets()
	if err != nil {
		fmt.Fprintf(w, "Error getting user sockets: %s", err.Error())
		return
	}
	var infoSlice []WebSocketInfo
	for key, userSocket := range userSockets {
		for _, socket := range userSocket {
			infoSlice = append(infoSlice, WebSocketInfo{RSSocket: socket, User_id: key, RemoteAddr: socket.Conn.RemoteAddr().String(), Subprotocol: socket.Conn.Subprotocol(), NetworkName: socket.Conn.LocalAddr().Network()})
		}
	}
	infoSliceJson, _ := json.Marshal(infoSlice)
	fmt.Fprintf(w, string(infoSliceJson))
}
