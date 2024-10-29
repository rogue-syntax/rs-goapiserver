package mainserver

import (
	"fmt"
	"math/big"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/rogue-syntax/rs-goapiserver/apierrors"

	"github.com/rogue-syntax/rs-goapiserver/apireturn/apierrorkeys"

	"runtime/debug"

	"github.com/rogue-syntax/rs-goapiserver/global"
	"github.com/rogue-syntax/rs-goapiserver/global/httpconfig"
)

// handler is a typical HTTP request-response handler in Go; details later
func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Greetings!")

}

func PanicRecovery(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, rq *http.Request) {

		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		rw.Header().Set("Access-Control-Allow-Credentials", "true")
		rw.Header().Set("Access-Control-Allow-Origin", "https://"+global.EnvVars.Apiserver+"")
		rw.Header().Set("Access-Control-Allow-Headers", "Content-Type,access-control-allow-origin, access-control-allow-headers")

		//start db conection to mariadb server : database.DB
		//err := database.StartDB()
		//if err != nil {
		//	apierrors.HandleError(nil, err, apierrorkeys.AppInitErr, &apierrors.ReturnError{Msg: apierrorkeys.AppInitErr, W: nil})
		//}

		defer func() {
			if err := recover(); err != nil {
				valStr := fmt.Sprint(err)
				panicErr := errors.New("Recovering from panic: " + valStr + " | stackTrace: " + string(debug.Stack()))
				if rw.Header().Get("Content-Type") == "" {
					rw.WriteHeader(http.StatusInternalServerError)
				}
				apierrors.HandleError(nil, panicErr, panicErr.Error(), &apierrors.ReturnError{Msg: apierrorkeys.PanicError, W: nil})
			}
		}()
		handler.ServeHTTP(rw, rq)
	})
}

func Serve() {
	fmt.Println("SERVING")

	seed := big.NewInt(time.Now().UnixNano())
	rand.New(rand.NewSource(seed.Int64()))

	//set http config
	httpconfig.SetHttpReqTimeout()
	http.HandleFunc("/v1/", handler)

	s := &http.Server{Addr: "0.0.0.0:9990", Handler: PanicRecovery(http.DefaultServeMux), ReadTimeout: 9600 * time.Second,
		WriteTimeout: 9600 * time.Second, IdleTimeout: 9600 * time.Second}
	l, err := net.Listen("tcp4", "0.0.0.0:9990")
	if err != nil {
		apierrors.HandleError(nil, err, err.Error(), &apierrors.ReturnError{Msg: apierrorkeys.ServeHttpError, W: nil})
	}
	s.Serve(l)
}
