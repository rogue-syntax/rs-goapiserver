package authentication

import (
	"context"
	"crypto/rand"
	"fmt"

	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/rogue-syntax/rs-goapiserver/apicontext"
	"github.com/rogue-syntax/rs-goapiserver/apierrors"
	"github.com/rogue-syntax/rs-goapiserver/apireturn"
	"github.com/rogue-syntax/rs-goapiserver/apireturn/apierrorkeys"
	"github.com/rogue-syntax/rs-goapiserver/authutil"
	"github.com/rogue-syntax/rs-goapiserver/database"
	"github.com/rogue-syntax/rs-goapiserver/entities/user"
	"github.com/rogue-syntax/rs-goapiserver/global"
	"github.com/rogue-syntax/rs-goapiserver/routeroles"
	"golang.org/x/crypto/bcrypt"
)

const (
	USER_ID_HEADER_KEY = "user-id"
)

/*

sign in with pw for user_auth_session issue token and get cookie
sign in with pw for user auth_session and get bearer token

user api tok in db should be compare encrypted:
user_auth_session issue token should be comapre encrypted:

Authenticate session with kbxs cookie , compare to user_auth_session.token
Authenticate session with kbxb , compare to user_auth_session.token
Authenticate one off request with api key in header kbxa , compare to user_api_tok

browser: user id cookie is kbxu, access token string is kbxs
non browser app: requests must be sent with kbxb header with token string. token string is returned to client at sign on and per validated request

*/

type HeaderAuthReturn struct {
	Kbxb string
}

type ApiKeyReturn struct {
	ApiSecret string
	Message   string
}

type UserSession struct {
	Sess_id      int `db:"Sess_id"`
	User_id      int
	Token        string
	Updated_at   int64
	Expires_at   int64
	User_agent   string
	User_ip_4    string
	User_ip_aton uint32
}

func AuthenticateUser(w *http.ResponseWriter, ctx context.Context) (*user.UserExternal, error) {
	usr, err := apicontext.CtxGetUser(ctx)
	if err != nil {
		apierrors.HandleError(nil, err, err.Error(), &apierrors.ReturnError{Msg: apierrorkeys.AuthorizationError, W: w})
		return nil, err
	}
	return usr, nil
}

// Handler_AppSignOut
//   - Expects json enocded ExternalUser
func Handler_AppSignOut(w http.ResponseWriter, r *http.Request, ctx context.Context) {

	usr, err := apicontext.CtxGetUser(ctx)

	if err != nil {

		apierrors.HandleError(nil, err, err.Error(), &apierrors.ReturnError{Msg: apierrorkeys.AuthorizationError, W: &w})
		return
	}

	err = killUserSessionForID_x_Agent(r, true, (*usr).User_id)
	if err != nil {
		apierrors.HandleError(nil, err, err.Error(), &apierrors.ReturnError{Msg: apierrorkeys.AuthorizationError, W: &w})
		return
	}

	apireturn.ApiJSONReturn(apierrorkeys.LoggedOut, apierrorkeys.NOError, &w)

}

func killUserSessionForID_x_Agent(r *http.Request, useCookie bool, user_id_in int) error {

	// lookup session based on user id and user agent
	// if request comes from non browser application this may ned to be set by application
	var user_id int
	var err error

	if useCookie == true {
		cookie, err := r.Cookie("kbxu")
		if err != nil {
			return err
		}
		user_id, err = strconv.Atoi(cookie.Value)
	} else {
		user_id, err = strconv.Atoi(r.FormValue(USER_ID_HEADER_KEY))
	}

	if err != nil {
		return err
	}
	//kbxu user_id cookie must match (*usr)
	if user_id == user_id_in {
		user_agent := authutil.Sha1Hash(r.Header.Get("User-Agent"))
		_, err = database.DB.DB.Exec("call killUserSession(?,?)", user_id, user_agent)
		if err != nil {
			return err
		}
	}
	return nil
}

// Handler_AppSignIn
//   - Sign in with post values 'pw, 'em', and 'kbxb'
//   - pw: the user's password
//   - em: the user's email
//   - kbxb: a non empty value specified for kbxb will result in a header token being returned for a non browser application to use for header authentication
//   - - an empty value will result in a samesite cookie being issued to the browser for browser app session authenication
//   - - conventipon for 'kbxb' will be the strings 'true', or the post body variable should be left unset
//   - - i.e. "kbxb: false" will result in a header token being retuned, just like "kbxb: true"
func Handler_AppSignIn(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	usr, err := verifyUser(r.FormValue("pw"), r.FormValue("em"))
	if err != nil {
		apierrors.HandleError(nil, err, err.Error(), &apierrors.ReturnError{Msg: apierrorkeys.AuthorizationError, W: &w})
		return
	}
	// password is authrnticated
	//issue token to cookie, or to header token
	isKbxb := r.FormValue("kbxb")
	userToken, err := issueToken((*usr).User_id, isKbxb, w, r)
	if err != nil {
		apierrors.HandleError(nil, err, err.Error(), &apierrors.ReturnError{Msg: apierrorkeys.AuthorizationError, W: &w})
		return
	}

	var ux user.UserExternal
	user.UserInternalExternal(usr, &ux)

	if isKbxb == "" {
		apireturn.ApiJSONReturn(ux, apierrorkeys.NOError, &w)
	} else {
		var authReturn HeaderAuthReturn
		authReturn.Kbxb = userToken
		apireturn.ApiJSONReturn(authReturn, apierrorkeys.NOError, &w)
	}

}

func HandleAppBrowserSignIn(pw string, em string, w http.ResponseWriter, r *http.Request) (*user.UserExternal, error) {
	var userExternal user.UserExternal
	usr, err := verifyUser(pw, em)
	if err != nil {
		return &userExternal, err
	}
	// password is authrnticated
	//issue token to cookie, or to header token
	isKbxb := r.FormValue("kbxb")
	_, err = issueToken((*usr).User_id, isKbxb, w, r)
	if err != nil {
		return &userExternal, err
	}

	user.UserInternalExternal(usr, &userExternal)
	return &userExternal, nil
}

func issueToken(user_id int, isKbxb string, w http.ResponseWriter, r *http.Request) (string, error) {
	var uSession UserSession

	bytesR := make([]byte, 16)
	rand.Read(bytesR)
	userToken := hex.EncodeToString(bytesR)
	uSession.Token = authutil.HashTokenBytes(bytesR)
	//give uSession.Token  encoded to client, store encrypted in db
	//not salting for now
	sExpiration := time.Now().Add(time.Duration(global.AuthTimeout) * time.Hour)
	uSession.Expires_at = sExpiration.Unix()
	uSession.Updated_at = time.Now().Unix()
	uSession.User_id = user_id
	uSession.User_agent = authutil.Sha1Hash(r.Header.Get("User-Agent"))
	uSession.User_ip_4 = authutil.ReadUserIP(r)
	uSession.User_ip_aton = authutil.InetAton(uSession.User_ip_4)

	if isKbxb == "" {
		//isKbxb value of 'kbxb' from request body not present so issue kookie
		issueSessionCookie(uSession.User_id, userToken, sExpiration, w)
	}
	err := UpdateUserSession(uSession)
	if err != nil {
		return userToken, err
	}
	return userToken, nil

}

func issueSessionCookie(user_id int, userToken string, sExpiration time.Time, w http.ResponseWriter) {
	cookie := http.Cookie{Name: "kbxs",
		Value:    userToken,
		Expires:  sExpiration,
		Domain:   global.EnvVars.Apiserver,
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}

	cookie2 := http.Cookie{Name: "kbxs",
		Value:    userToken,
		Expires:  sExpiration,
		Domain:   "www." + global.EnvVars.Apiserver,
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}

	cookie3 := http.Cookie{Name: "kbxu",
		Value:    strconv.Itoa(user_id),
		Expires:  sExpiration,
		Domain:   global.EnvVars.Apiserver,
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}

	cookie4 := http.Cookie{Name: "kbxu",
		Value:    strconv.Itoa(user_id),
		Expires:  sExpiration,
		Domain:   "www." + global.EnvVars.Apiserver,
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, &cookie)
	http.SetCookie(w, &cookie2)
	http.SetCookie(w, &cookie3)
	http.SetCookie(w, &cookie4)
}

func UpdateUserSession(uSession UserSession) error {
	_, err := database.DB.Exec("call UpdateUserSession(?,?,?,?,?,?,?)",
		uSession.User_id,
		uSession.Token,
		uSession.Updated_at,
		uSession.Expires_at,
		uSession.User_agent,
		uSession.User_ip_4,
		uSession.User_ip_aton,
	)
	if err != nil {
		return err
	}
	return nil

}

// PW Verif
//   - Use bcrypt to compare submitted password and hased store
//   - returns boolean true for positive match, flas for negative match, and a checkable error
func pwVerif(hString *string, pwIn *string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(*hString), []byte(*pwIn))
	if err != nil {
		return false, err
	}
	return true, err
}

// Verify User
//
//   - Param: pw string - post request submitted password
//   - Param: em string - post request submitted email
//   - Returns: *user.UserInternal, error
//   - Attempts to find user by email using user.FindUserInternalByEmail
//   - Attempts to get positive comparision between submitted pw and hashed pw from database user record
//   - Will either return a non nil error, or a user.UserInternal object
func verifyUser(pw string, em string) (*user.UserInternal, error) {
	usr, err := user.FindUserInternalByEmail(em)
	if err != nil {
		return usr, err
	}
	isAuthentic, err := pwVerif(&usr.User_pw, &pw)
	if err != nil {
		return usr, err
	}
	if isAuthentic == true {
		return usr, nil
	}
	err = errors.New(apierrorkeys.PWIncorrect)
	return usr, err
}

func Handler_GenApiKey(w http.ResponseWriter, r *http.Request, ctx context.Context) {

	usr, err := apicontext.CtxGetUser(ctx)

	if err != nil {
		apierrors.HandleError(nil, err, err.Error(), &apierrors.ReturnError{Msg: apierrorkeys.AuthorizationError, W: &w})
		return
	}

	hexStrForClient, bytesForDB, err := authutil.MakeAuthToken()
	if err != nil {
		apierrors.HandleError(nil, err, err.Error(), &apierrors.ReturnError{Msg: apierrorkeys.AuthorizationError, W: &w})
		return
	}
	//return hex string to user
	//save sha256 hashed -> hex string to database

	hashForStorage := authutil.HashTokenBytes(bytesForDB)
	_, err = database.DB.Exec("call UpdateUserApiTok(?,?)", (*usr).User_id, hashForStorage)
	if err != nil {
		apierrors.HandleError(nil, err, err.Error(), &apierrors.ReturnError{Msg: apierrorkeys.AuthorizationError, W: &w})
		return
	}
	apiKeyReturn := ApiKeyReturn{ApiSecret: hexStrForClient,
		Message: "Api secret attached. You are responsible for any action taken on your behalf with this key. Do not loose it. Do not share it."}

	apireturn.ApiJSONReturn(apiKeyReturn, apierrorkeys.NOError, &w)
	return
}

func generateRandomSalt(saltSize int) ([]byte, error) {
	var salt = make([]byte, saltSize)
	_, err := rand.Read(salt[:])
	if err != nil {
		return salt, err
	}
	return salt, nil

}

func Handler_TestReqVerif(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	usr, err := apicontext.CtxGetUser(ctx)

	if err != nil {
		apierrors.HandleError(nil, err, err.Error(), &apierrors.ReturnError{Msg: apierrorkeys.AuthorizationError, W: &w})
		return
	}

	apireturn.ApiJSONReturn(usr, apierrorkeys.NOError, &w)
}

func compareAndVerify(user_id int, userToken string, hasedToken string, ctx context.Context) (context.Context, error) {
	err := tokenCompare(userToken, hasedToken)
	if err != nil {
		return ctx, err
	}

	usr, err := user.FindUserExternalByUser_id(user_id)
	if err != nil {
		return ctx, err
	}

	ctx = apicontext.CtxWithUser(ctx, usr)
	return ctx, nil
}

// Verify with api
//   - Branched from VerifyRequest
func VerifyWithApi(ctx context.Context, routeString string, w http.ResponseWriter, r *http.Request) (context.Context, error) {
	apiKey := r.Header.Get("kbxa")
	if apiKey != "" {
		//user_id, err := strconv.Atoi(r.FormValue(USER_ID_HEADER_KEY))
		user_id, err := strconv.Atoi(r.Header.Get(USER_ID_HEADER_KEY))
		if err != nil {
			return ctx, err
		}
		apiKeyHash, err := user.FindApiKeyByUser_id(user_id)
		if err != nil {
			return ctx, err
		}
		ctx, err := compareAndVerify(user_id, apiKey, apiKeyHash, ctx)
		return ctx, err
	} else {
		err := errors.New(apierrorkeys.APIKeyNotFound)
		return ctx, err
	}
}

// Verify with header
//   - Branched from VerifyRequest
func VerifyWithHeader(ctx context.Context, routeString string, w http.ResponseWriter, r *http.Request) (context.Context, error) {
	apiKey := r.Header.Get("kbxb")
	if apiKey != "" {
		uSession, err := getUserSessionForID_x_Agent(r, false)
		if err != nil {
			return ctx, err
		}
		ctx, err = compareAndVerify(uSession.User_id, apiKey, uSession.Token, ctx)
		return ctx, err
	} else {
		err := errors.New(apierrorkeys.APIKeyNotFound)
		return ctx, err
	}
}

// Verify with cookie
//   - Branched from VerifyRequest
func VerifyWithCookie(ctx context.Context, routeString string, w http.ResponseWriter, r *http.Request) (context.Context, error) {
	apiKey := ""
	cookie, cookErr := r.Cookie("kbxs")
	if cookErr != nil {
		err := errors.New(apierrorkeys.AuthorizationError)
		return ctx, err
	} else {
		//cookie found
		apiKey = cookie.Value
	}
	uSession, err := getUserSessionForID_x_Agent(r, true)
	if err != nil {
		return ctx, err
	}

	ctx, err = compareAndVerify(uSession.User_id, apiKey, uSession.Token, ctx)
	if err != nil {
		return ctx, err
	}

	//dont reissue on every request to avoid sync issues
	/*
		_, err = issueToken(uSession.User_id, "", w, r)
	*/
	return ctx, nil

}

// VERIFY REQUEST
// Request will be verified one od three ways:
//   - Authenticate session with kbxs cookie , compare to user_auth_session.token
//   - - requires post body : user_id and header: user agent
//   - Authenticate session with kbxb , compare to user_auth_session.token
//   - - Purpose of VerifyWithHeader with authBode b : kbxb header is to provide header api auth that requires login session to be active, i.e. a native application that requires login but does not store cookies.
//   - - i.e. VerifyWithHeader checks to see if login session is outdated or not, requires a login session timestamp to not be expired
//   - - requires post body : authMode "b"
//   - - requires post body : user_id and header: user agent
//   - Authenticate one off request with api key in header kbxa , compare to user_api_tok
//   - - Purpose of VerifyWithAPI with authBode a : kbxa header isto provide per prequest API authentication for third parties accessing data via API RPC calls
//   - - requires post body : authMode "a"
func VerifyRequest(ctx context.Context, routeString string, authMode string, w http.ResponseWriter, r *http.Request) (context.Context, error) {
	if authMode == "a" {
		ctx, err := VerifyWithApi(ctx, routeString, w, r)
		errx := errors.Wrap(err, apierrorkeys.AuthorizationError)
		return ctx, errx
	} else if authMode == "b" {
		ctx, err := VerifyWithHeader(ctx, routeString, w, r)
		errx := errors.Wrap(err, apierrorkeys.AuthorizationError)
		return ctx, errx
	} else {
		ctx, err := VerifyWithCookie(ctx, routeString, w, r)
		errx := errors.Wrap(err, apierrorkeys.AuthorizationError)
		return ctx, errx
	}
}

func tokenCompare(userToken string, sessionToken string) error {

	tBytes, err := hex.DecodeString(userToken)
	if err != nil {
		return err
	}
	hashedToken := authutil.HashTokenBytes(tBytes)

	if hashedToken != sessionToken {
		//err = errors.New("userToken: " + userToken + " / hashedToken: " + hashedToken + " / session: " + sessionToken)
		err = errors.New(apierrorkeys.AuthorizationError)
		return err
	}
	return nil
	/*
		//method 2: constant time compare
		tokenBytes, err := hex.DecodeString(token)
		sessionTokenBytes, err := hex.DecodeString(uSession.Token)
		if err != nil {
			return ctx, err
		}
		if subtle.ConstantTimeCompare(tokenBytes, sessionTokenBytes) != 1 {
			err = errors.New(apierrorkeys.AuthorizationError)
			return ctx, err
		}
	*/
}

func GetUserAgentHashFromRequest(r *http.Request) string {
	user_agent := authutil.Sha1Hash(r.Header.Get("User-Agent"))
	return user_agent
}

func getUserSessionForID_x_Agent(r *http.Request, useCookie bool) (UserSession, error) {
	var uSession UserSession
	// lookup session based on user id and user agent
	// if request comes from non browser application this may ned to be set by application
	var user_id int
	var err error
	if useCookie == true {
		cookie, err := r.Cookie("kbxu")
		if err != nil {
			return uSession, err
		}
		user_id, err = strconv.Atoi(cookie.Value)
	} else {
		user_id, err = strconv.Atoi(r.FormValue(USER_ID_HEADER_KEY))
	}

	if err != nil {
		return uSession, err
	}

	user_agent := authutil.Sha1Hash(r.Header.Get("User-Agent"))

	err = database.DB.Get(&uSession, "call getUserSession(?,?)", user_id, user_agent)
	if err != nil {
		return uSession, err
	}
	// is thie session expired
	if int(uSession.Expires_at) < int(time.Now().Unix()) {
		sessionErr := errors.New(apierrorkeys.SessionExpired)
		return uSession, sessionErr
	}

	return uSession, nil

}

func MapSliceContains(s []int, roleId int) bool {
	for _, v := range s {
		if v == roleId {
			return true
		}
	}
	return false
}

func HasRoutePermission(roleId int, resource string) bool {
	fmt.Printf("resource:  %s \r\n", string(resource))

	if !MapSliceContains(routeroles.RouteRoles[resource], roleId) {
		return false
	} else {
		return true
	}
}
