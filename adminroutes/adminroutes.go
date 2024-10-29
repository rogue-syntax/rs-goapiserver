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

func SetAdminRoutes() {

	middleware.RouteHandler("/v1/test/gen-pw", func(w http.ResponseWriter, r *http.Request, ctx context.Context) {

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

		//htmlStr := "<div>" + dStr + "</div>"
		//htmlStr += "<script>let errorArray = " + dStr + "; console.log(errorArray[errorArray.length-2]);console.log(errorArray)</script>"

		htmlStr := "<div>" + dStr + "</div>"
		htmlStr += "<script>let errorArray = " + dStr + ";"
		htmlStr += "errorArray.forEach((err, i) => {"
		htmlStr += "		try{ errorArray[i].context = JSON.parse(errorArray[i].error); }catch(e){};"
		htmlStr += " }); console.log(errorArray[errorArray.length-2]);  console.log(errorArray)</script>"

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, htmlStr)

	}, &middleware.BlankMiddleware)

	middleware.RouteHandler("/v1/errors/reqLog", func(w http.ResponseWriter, r *http.Request, ctx context.Context) {
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

		//htmlStr := "<div>" + dStr + "</div>"
		//htmlStr += "<script>let errorArray = " + dStr + "; console.log(errorArray[errorArray.length-2]);console.log(errorArray)</script>"

		htmlStr := "<div>" + dStr + "</div>"
		htmlStr += "<script>let logArray = " + dStr + ";"
		//htmlStr += "function superParse( obj ){ Object.keys(obj).forEach( (key) =>{ success = false; try{ obj[key] = JSON.parse(obj[key]); success = true; }catch(e){}; if ( success === true ){ superParse(obj[key])}; })};"
		//htmlStr += "logArray.forEach((err, i) => {"
		//htmlStr += "		try{ logArray[i].context = JSON.parse(logArray[i].log); }catch(e){};"
		//htmlStr += " }); console.log(logArray[logArray.length-2]);"
		htmlStr += "console.log(logArray)</script>"

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, htmlStr)

	}, &middleware.BlankMiddleware)

	middleware.RouteHandler("/v1/test/testPWVerificationEP", signup.TestPWVerifEP_handler, &middleware.BlankMiddleware)

	middleware.RouteHandler("/v1/test/testRoleAuthentication", routeroles.TestRoleAuth, &middleware.RoleBaseReqVerifMiddleware)

	middleware.RouteHandler("/v1/observe/logGoroutineCount", observability.Handler_LogGoroutineCount, &middleware.RoleBaseReqVerifMiddleware)

	middleware.RouteHandler("/v1/observe/getUserSockets", observability.Handler_GetUserSockets, &middleware.RoleBaseReqVerifMiddleware)

}
