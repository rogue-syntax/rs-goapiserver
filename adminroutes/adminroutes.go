package adminroutes

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/rogue-syntax/rs-goapiserver/authutil"
	"github.com/rogue-syntax/rs-goapiserver/middleware"
	"github.com/rogue-syntax/rs-goapiserver/observability"
	"github.com/rogue-syntax/rs-goapiserver/routeroles"

	"github.com/rogue-syntax/rs-goapiserver/signup"
)

// Admin route handlers
func Handler_GenPW(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	pwStr := r.FormValue("pw")
	pwHash, _ := authutil.GeneratePW(pwStr)
	fmt.Fprint(w, pwHash)
}

func Handler_ErrLog(w http.ResponseWriter, r *http.Request, ctx context.Context) {
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
	htmlStr += "<script>let errorArray = " + dStr + ";"
	htmlStr += "errorArray.forEach((err, i) => {"
	htmlStr += "		try{ errorArray[i].context = JSON.parse(errorArray[i].error); }catch(e){};"
	htmlStr += " }); console.log(errorArray[errorArray.length-2]);  console.log(errorArray)</script>"

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, htmlStr)
}

func Handler_ReqLog(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	data, _ := os.ReadFile("/var/logs/requestLogger.log")
	fmt.Fprintf(w, string(data))
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
	htmlStr += "<script>let logArray = " + dStr + ";"
	htmlStr += "console.log(logArray)</script>"

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, htmlStr)
}

// Route definitions
var AdminRoutes = []middleware.RouteDef{
	{
		RouteStr:       "/v1/test/gen-pw",
		HandlerFunc:    Handler_GenPW,
		MiddlewareSli:  &middleware.BlankMiddleware,
		PermittedRoles: []int{0, 1}, // Billing and Account Admin only
	},
	{
		RouteStr:       "/v1/errors/errLog",
		HandlerFunc:    Handler_ErrLog,
		MiddlewareSli:  &middleware.BlankMiddleware,
		PermittedRoles: []int{0, 1}, // Billing and Account Admin only
	},
	{
		RouteStr:       "/v1/errors/reqLog",
		HandlerFunc:    Handler_ReqLog,
		MiddlewareSli:  &middleware.BlankMiddleware,
		PermittedRoles: []int{0, 1}, // Billing and Account Admin only
	},
	{
		RouteStr:       "/v1/test/testPWVerificationEP",
		HandlerFunc:    signup.TestPWVerifEP_handler,
		MiddlewareSli:  &middleware.BlankMiddleware,
		PermittedRoles: []int{0, 1}, // Billing and Account Admin only
	},
	{
		RouteStr:       "/v1/test/testRoleAuthentication",
		HandlerFunc:    routeroles.TestRoleAuth,
		MiddlewareSli:  &middleware.RoleBaseReqVerifMiddleware,
		PermittedRoles: []int{0, 1}, // Billing and Account Admin only
	},
	{
		RouteStr:       "/v1/observe/logGoroutineCount",
		HandlerFunc:    observability.Handler_LogGoroutineCount,
		MiddlewareSli:  &middleware.RoleBaseReqVerifMiddleware,
		PermittedRoles: []int{0, 1}, // Billing and Account Admin only
	},
	{
		RouteStr:       "/v1/observe/getUserSockets",
		HandlerFunc:    observability.Handler_GetUserSockets,
		MiddlewareSli:  &middleware.RoleBaseReqVerifMiddleware,
		PermittedRoles: []int{0, 1}, // Billing and Account Admin only
	},
}

func SetAdminRoutes() {
	middleware.SetRouteDefs(&AdminRoutes, "admin")
}
