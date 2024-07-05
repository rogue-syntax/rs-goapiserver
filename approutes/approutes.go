package approutes

import (
	"github.com/rogue-syntax/rs-goapiserver/apimaster"
	"github.com/rogue-syntax/rs-goapiserver/authentication"
	"github.com/rogue-syntax/rs-goapiserver/mail"
	"github.com/rogue-syntax/rs-goapiserver/middleware"
	"github.com/rogue-syntax/rs-goapiserver/signup"
	"github.com/rogue-syntax/rs-goapiserver/websockets"
)

var BaseAppRoutes = []middleware.RouteDef{
	{RouteStr: "/v1/api", HandlerFunc: apimaster.Handler_GetApiReqMapPage, MiddlewareSli: &middleware.RoleBaseReqVerifMiddleware},
	{RouteStr: "/v1/api-data", HandlerFunc: apimaster.Handler_GetApiReqMap, MiddlewareSli: &middleware.RoleBaseReqVerifMiddleware},
	{RouteStr: "/v1/app/signIn", HandlerFunc: authentication.Handler_AppSignIn, MiddlewareSli: &middleware.BlankMiddleware},
	{RouteStr: "/v1/app/signup", HandlerFunc: signup.Handler_AppSignUp, MiddlewareSli: &middleware.BlankMiddleware},
	{RouteStr: "/v1/app/signOut", HandlerFunc: authentication.Handler_AppSignOut, MiddlewareSli: &middleware.ReqVerifMiddleware},
	{RouteStr: "/v1/app/testReqVerif", HandlerFunc: authentication.Handler_TestReqVerif, MiddlewareSli: &middleware.ReqVerifMiddleware},
	{RouteStr: "/v1/app/testEmail", HandlerFunc: mail.SendTestEmail_handler, MiddlewareSli: &middleware.BlankMiddleware},
	{RouteStr: "/v1/app/emailVerificationEP", HandlerFunc: signup.EmailVerifEP_handler, MiddlewareSli: &middleware.BlankMiddleware},
	{RouteStr: "/v1/app/requestPasswordReset", HandlerFunc: signup.Handler_RequestPasswordReset, MiddlewareSli: &middleware.BlankMiddleware},
	{RouteStr: "/v1/app/newPWVerificationEP", HandlerFunc: signup.PWVerifEP_handler, MiddlewareSli: &middleware.BlankMiddleware, ReqDef: &signup.PWVerifEP_handler_ApiReq},
	{RouteStr: "/v1/testWS/", HandlerFunc: websockets.TestWS, MiddlewareSli: &middleware.ReqVerifMiddleware},
	{RouteStr: "/ws/wss/", HandlerFunc: websockets.WsEndpoint, MiddlewareSli: &middleware.ReqVerifMiddleware},
	{RouteStr: "/v1/test/genApiKey", HandlerFunc: authentication.Handler_GenApiKey, MiddlewareSli: &middleware.ReqVerifMiddleware},
}
