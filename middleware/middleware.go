package middleware

import (
	"bytes"
	"context"
	"io"

	"net/http"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/rogue-syntax/rs-goapiserver/apicontext"
	"github.com/rogue-syntax/rs-goapiserver/apierrors"
	"github.com/rogue-syntax/rs-goapiserver/apimaster"
	"github.com/rogue-syntax/rs-goapiserver/rs_go_requestlogger"

	"github.com/rogue-syntax/rs-goapiserver/authentication"
)

// A Middleware wrapper for HTTP / Net package
/*

	A 'middlware' here is a collection of ProcessRequest functions to execute prior to the eventual api route handler: []type<RequestMiddleware>.ProcessRequest()

	For each of the middlware processes to be integrated into this middleware,
	create a type that will fulfil the RequestMiddleware interface by implementing a ProcessRequest method with the signature:
	(ctx context.Context, routeString string, w http.ResponseWriter, r *http.Request) (context.Context, error)

	instatiate an object of each type

	instatiate a slice of these as []RequestMiddleware

	create the setMiddleMidleware function to Push each of those into a []RequestMiddleware

	Define the logic for each type<RequestMiddleware> ProcessRequest ()

	in the api definition, include the middeware slice that contains the middlware processes needed for that route:
	middleware.RouteHandler("/v1/app/validatePhoneNumber", onboarding.Handler_validatePhoneNumber, &middleware.ReqVerifMiddleware)
	"url route, handler function, middleware slice"

	when the RouteHandler middlware wrapper executes the prequested url endpoint,
	the HTTP / Net package http.HandleFunc handler will iterate through the provided middleware slice and run the ProcessRequest function of each type<RequestMiddleware>,
	passing a context object to and from each function call, and ultimately the route handling function.

	If no errors are encountered, execution is handed off to the request handler specified in the route definition

	If non nil errors are encountered, they are handled by the apierrors.HandleRouterErrors handler,
	and the process immediately returns and terminates prior to handing off execution to the eventual route handler.

*/

const (
	AUTH_MODE_KEY = "auth-mode"
)

/*
Eventual Handler: Signtaure type for route handling function

-	To be eventually called after middleware ProcessRequest(s) have been executed
*/
type EventualHandler func(http.ResponseWriter, *http.Request, context.Context)

// RouteHandler
//   - takes the url route, http request and writer
//   - processes them though a slice of RequestMiddleware using its ProcessRequest function
//   - passes a requestContext though these middleware functions, carryting a context object through to our requesthandlers
//   - example : RouteHandler("/v1/someRoute", SomeFunction, [] )
func RouteHandler(routeString string, reqHandler EventualHandler, mwList *[]RequestMiddleware) {
	http.HandleFunc(routeString, func(w http.ResponseWriter, r *http.Request) {
		reqCtx := context.Background()
		//LOG REQUEST HERE
		var rSRequestLogger rs_go_requestlogger.RSRequestLogger
		rSRequestLogger.Endpoint = routeString

		rSRequestLogger.RequestVars.RequestURI = r.RequestURI
		rSRequestLogger.RequestVars.Method = r.Method
		rSRequestLogger.RequestVars.RemoteAddr = r.RemoteAddr
		rSRequestLogger.RequestVars.UserAgent = r.UserAgent()
		rSRequestLogger.RequestVars.Referer = r.Referer()
		rSRequestLogger.RequestVars.Host = r.Host
		rSRequestLogger.RequestVars.Header = r.Header

		//GET FORM DATA
		r.ParseForm()
		rSRequestLogger.RequestVars.PostForm = r.PostForm

		//GET COOKIES
		cookies := r.Cookies()
		cookieMap := make(map[string]string)
		for _, cookie := range cookies {
			cookieMap[cookie.Name] = cookie.Value
		}
		rSRequestLogger.RequestVars.Cookies = cookieMap
		rSRequestLogger.RequestVars.URL = r.URL.String()

		//GET BODY IN STRING FORM
		bytedata, _ := io.ReadAll(r.Body)
		reqBodyString := string(bytedata)
		rSRequestLogger.RequestVars.Body = reqBodyString
		r.Body = io.NopCloser(bytes.NewBuffer(bytedata))

		reqID := uuid.New().String()
		rSRequestLogger.Req_id = reqID
		//STORE LOG IN CTX
		reqCtx = rs_go_requestlogger.CtxWithRSLogger(reqCtx, &rSRequestLogger)

		reqCtx = rs_go_requestlogger.CtxWithReqId(reqCtx, reqID)

		r = r.WithContext(reqCtx)

		var err error
		for i := 0; i < len(*mwList); i++ {
			reqCtx, err = (*mwList)[i].ProcessRequest(reqCtx, routeString, w, r)
			if err != nil {
				apierrors.HandleError(nil, err, err.Error(), &apierrors.ReturnError{Msg: err.Error(), W: &w})
				return
			}
		}
		//request logger on request finished
		defer func() {
			apierrors.HandleReqLog(r)
		}()
		reqHandler(w, r, reqCtx)
	})
}

/*
RouteDef: A Struct for defining routes
  - RouteStr: The endpoint where route can be reached , i.e. "/v1/getSomething"
  - HandlerFunc: The handling fucntion to contain the business logic of the route
  - MiddlewareSli: The slice collection of middleware obects implementing the RequestMiddleware interface,
    containing the ProcessRequest functions that contain the middlware logic
*/
type RouteDef struct {
	RouteStr      string
	HandlerFunc   EventualHandler
	MiddlewareSli *[]RequestMiddleware
	ReqDef        *apimaster.ApiReqDef
}

/*
SetRouteDefs:
  - SetRouteDefs takes a pointer to a slice collection of RouteDef route definitions,
    and registers them with the RouteHandler function of the middleware package
  - Example: \
    var routeDef RouteDef{RouteStr:}
*/

func SetRouteDefs(defs *[]RouteDef, listName string) {
	//routeDef := RouteDef{RouteStr: "/v1/postSomething", }
	//if the listName / category doesn't exist yet, create it
	_, ok := apimaster.ApiReqMap[listName]
	if !ok {
		apimaster.ApiReqMap[listName] = make(map[string]apimaster.ApiReqDef)
	}
	for _, def := range *defs {
		//register the route with the middleware
		RouteHandler(def.RouteStr, def.HandlerFunc, def.MiddlewareSli)
		//register the route with the apimaster api request map
		if def.ReqDef != nil {
			apimaster.ApiReqMap[listName][def.RouteStr] = *def.ReqDef
		}
	}
}

// Request Middleware
// - interface for all middleware handlers
type RequestMiddleware interface {
	ProcessRequest(ctx context.Context, routeString string, w http.ResponseWriter, r *http.Request) (context.Context, error)
}

// DEV MIDDLEWARE
//   - 1. <ake a Middleware type i.e. type ExampleMiddlewareOneType struct
//   - 2. Middleware declare an object of that type i.e. var ExampleMiddlewareOne ExampleMiddlewareOneType
//   - 3. Declare a list of middleware handlers of []RequestMiddleware
//   - 4. Create set middlerare function to call from main
//   - 5. Use that function to create your []RequestMiddleware i.e. ExampleMiddleware = []RequestMiddleware{&ExampleMiddlewareOne, &ExampleMiddlewareTwo}
//   - 6. Attach a function to your middleware type that implenets the RequestMiddleware interfacew i.e. func (ExampleMiddlewareOneType) ProcessRequest ( ctx, routeString... )
//   - 7. Assign middleware to a route somewhere in main i.e. middleware.RouteHandler("/v1/test/genAuthToken", authentication.Handler_AppSignIn, &middleware.SignInMiddleware)
type ExampleMiddlewareOneType struct {
}

var ExampleMiddlewareOne ExampleMiddlewareOneType

type ExampleMiddlewareTwoType struct {
}

var ExampleMiddlewareTwo ExampleMiddlewareTwoType

var ExampleMiddleware []RequestMiddleware

func SetExampleMiddleware() {
	ExampleMiddleware = []RequestMiddleware{&ExampleMiddlewareOne, &ExampleMiddlewareTwo}
}

func (ExampleMiddlewareOneType) ProcessRequest(ctx context.Context, routeString string, w http.ResponseWriter, r *http.Request) (context.Context, error) {
	str := "User Terry"
	ctx = apicontext.CtxWithdevGenMsg(ctx, &str)
	return ctx, nil
}

func (ExampleMiddlewareTwoType) ProcessRequest(ctx context.Context, routeString string, w http.ResponseWriter, r *http.Request) (context.Context, error) {
	str := "User Bob"
	ctx = apicontext.CtxWithdevGenMsg(ctx, &str)
	return ctx, nil
}

// //////////////////
// BLANK MIDDLEWARE
type BlankMWType struct {
}

var BlankMW BlankMWType
var BlankMiddleware []RequestMiddleware

func SetBlankMiddleware() {
	BlankMiddleware = []RequestMiddleware{&BlankMW}
}
func (BlankMWType) ProcessRequest(ctx context.Context, routeString string, w http.ResponseWriter, r *http.Request) (context.Context, error) {
	var err error
	return ctx, err
}

////////////////
//APP MIDDLEWARE

// //////////////////////
// REQUEST VERIFICATION
//   - Get User in handler with : usr, err := apicontext.CtxGetUser(ctx)
type RequestVerifType struct {
}

var ReqVerif RequestVerifType
var ReqVerifMiddleware []RequestMiddleware

func SetReqVerifMiddleware() {
	ReqVerifMiddleware = []RequestMiddleware{&ReqVerif}
}
func (RequestVerifType) ProcessRequest(ctx context.Context, routeString string, w http.ResponseWriter, r *http.Request) (context.Context, error) {
	ctx, err := authentication.VerifyRequest(ctx, routeString, r.FormValue(AUTH_MODE_KEY), w, r)
	usr, _ := apicontext.CtxGetUser(ctx)
	r = r.WithContext(apicontext.CtxWithUser(r.Context(), usr))
	return ctx, err
}

/*
	Example operation to get user from authentication middleware context
	usr, err := apicontext.CtxGetUser(ctx)
	if err != nil {
		apierrors.HandleApiReqErrors(err, apiErrorKeys.APIReqError+": Handler_GetDataPointObjects", nil, w, r)
	}
*/

// /////////////////////
// WEBHOOK MW
type WebhookMWType struct {
}

var WebhookMW WebhookMWType
var WebhookMiddleware []RequestMiddleware

func SetWebhookMiddleware() {
	WebhookMiddleware = []RequestMiddleware{&WebhookMW}
}
func (WebhookMWType) ProcessRequest(ctx context.Context, routeString string, w http.ResponseWriter, r *http.Request) (context.Context, error) {
	//var err error
	//verify webhook authenticity here or at handlers?
	//ctx, err := authentication.VerifyRequest(ctx, routeString, r.FormValue("authMode"), w, r)
	return ctx, nil
}

// //////////////////////////////////
// ROLE BASED REQUEST VERIFICATION
type RoleBasedRequestVerifType struct {
}

var RoleBaseReqVerif RoleBasedRequestVerifType

var RoleBaseReqVerifMiddleware []RequestMiddleware

func SetRoleBaseReqVerifMiddleware() {
	RoleBaseReqVerifMiddleware = []RequestMiddleware{&RoleBaseReqVerif}
}

func (RoleBasedRequestVerifType) ProcessRequest(ctx context.Context, routeString string, w http.ResponseWriter, r *http.Request) (context.Context, error) {
	ctx, err := authentication.VerifyRequest(ctx, routeString, r.FormValue(AUTH_MODE_KEY), w, r)
	if err != nil {
		return ctx, err
	}

	usr, err := apicontext.CtxGetUser(ctx)
	if err != nil {
		return ctx, err
	}

	hasPermission := authentication.HasRoutePermission(*usr.User_role_id, routeString)
	if hasPermission != true {
		err = errors.New("User not authenticated for this route")
	}
	return ctx, err
}
