package adminroutes

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"rs-apiserver.com/authutil"
	"rs-apiserver.com/middleware"
	"rs-apiserver.com/routeroles"

	"rs-apiserver.com/signup"
)

func SetAdminRoutes() {

	middleware.RouteHandler("/v1/test/genPW", func(w http.ResponseWriter, r *http.Request, ctx context.Context) {
		pwStr := r.FormValue("pw")
		pwHash, _ := authutil.GeneratePW(pwStr)
		fmt.Fprint(w, pwHash)
	}, &middleware.BlankMiddleware)

	middleware.RouteHandler("/v1/errors/errLog", func(w http.ResponseWriter, r *http.Request, ctx context.Context) {
		data, _ := os.ReadFile("/var/logs/apiserver.log")
		dstr := string(data)
		dstrSli := strings.Split(dstr, "\n")
		dStr := "["
		for i, strs := range dstrSli {
			if i != 0 || i != len(dstrSli)-1 {
				dStr += strs + ","
			}

		}
		dStr += "]"

		htmlStr := "<div>" + dStr + "</div>"
		htmlStr += "<script>let errorArray = " + dStr + "; console.log(errorArray)</script>"
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, htmlStr)

	}, &middleware.BlankMiddleware)

	middleware.RouteHandler("/v1/test/testPWVerificationEP", signup.TestPWVerifEP_handler, &middleware.BlankMiddleware)

	middleware.RouteHandler("/v1/test/testRoleAuthentication", routeroles.TestRoleAuth, &middleware.RoleBaseReqVerifMiddleware)

}
