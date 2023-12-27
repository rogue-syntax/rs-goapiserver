package approutes

import (
	"github.com/rogue-syntax/rs-goapiserver/authentication"
	"github.com/rogue-syntax/rs-goapiserver/mail"
	"github.com/rogue-syntax/rs-goapiserver/middleware"
	"github.com/rogue-syntax/rs-goapiserver/signup"
	"github.com/rogue-syntax/rs-goapiserver/websockets"
)

func SetAppRoutes() {

	middleware.RouteHandler("/v1/app/signIn", authentication.Handler_AppSignIn, &middleware.BlankMiddleware)

	middleware.RouteHandler("/v1/app/signup", signup.Handler_AppSignUp, &middleware.BlankMiddleware)

	middleware.RouteHandler("/v1/app/signOut", authentication.Handler_AppSignOut, &middleware.ReqVerifMiddleware)

	middleware.RouteHandler("/v1/app/testReqVerif", authentication.Handler_TestReqVerif, &middleware.ReqVerifMiddleware)

	middleware.RouteHandler("/v1/app/testEmail", mail.SendTestEmail_handler, &middleware.BlankMiddleware)

	middleware.RouteHandler("/v1/app/emailVerificationEP", signup.EmailVerifEP_handler, &middleware.BlankMiddleware)

	middleware.RouteHandler("/v1/app/requestPasswordReset", signup.Handler_RequestPasswordReset, &middleware.BlankMiddleware)

	middleware.RouteHandler("/v1/app/newPWVerificationEP", signup.PWVerifEP_handler, &middleware.BlankMiddleware)

	middleware.RouteHandler("/v1/testWS/", websockets.TestWS, &middleware.ReqVerifMiddleware)

	middleware.RouteHandler("/ws/wss/", websockets.WsEndpoint, &middleware.ReqVerifMiddleware)

	middleware.RouteHandler("/v1/test/genApiKey", authentication.Handler_GenApiKey, &middleware.ReqVerifMiddleware)

}
