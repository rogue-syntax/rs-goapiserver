package httpconfig

import (
	"net/http"
	"time"
)

func SetHttpReqTimeout() {
	http.DefaultTransport.(*http.Transport).ResponseHeaderTimeout = time.Second * 60
}
