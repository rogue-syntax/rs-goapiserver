package mainserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rogue-syntax/rs-goapiserver/apierrors"

	"github.com/rogue-syntax/rs-goapiserver/apireturn/apierrorkeys"
	"github.com/rogue-syntax/rs-goapiserver/database"

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

func Serve() (context.Context, *http.ServeMux) {

	root, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Global HTTP timeouts/config
	httpconfig.SetHttpReqTimeout()

	// Start DB ONCE at startup, close on shutdown
	if err := database.StartDB(root); err != nil {
		apierrors.HandleError(nil, err, err.Error(), &apierrors.ReturnError{Msg: apierrorkeys.SendMailError, W: nil})
		return root, nil
	}
	defer func() {
		// If you have a Close() or Stop() in your database pkg, call it here.
		if closer, ok := interface{}(database.DB).(interface{ Close() error }); ok && database.DB != nil {
			_ = closer.Close()
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/", handler)

	srv := &http.Server{
		Addr:    "0.0.0.0:9990",
		Handler: PanicRecovery(mux),
		// Reasonable timeouts (9600s is… a lot)
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
		// Inherit root context so every request gets canceled on shutdown
		BaseContext: func(net.Listener) context.Context { return root },
	}

	ln, err := net.Listen("tcp4", srv.Addr)
	if err != nil {
		apierrors.HandleError(nil, err, err.Error(), &apierrors.ReturnError{Msg: apierrorkeys.ServeHttpError, W: nil})
		return root, nil
	}

	// Run the server
	serverErr := make(chan error, 1)
	go func() {
		// When Shutdown begins, Serve returns http.ErrServerClosed – not an error
		if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	// Wait for signal or server error
	select {
	case <-root.Done():
		// graceful shutdown
	case err := <-serverErr:
		if err != nil {
			apierrors.HandleError(nil, err, err.Error(), &apierrors.ReturnError{Msg: apierrorkeys.ServeHttpError, W: nil})
		}
	}

	// Give in-flight requests time to finish
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		// Force-close if graceful shutdown times out
		_ = srv.Close()
	}

	return root, mux
}
