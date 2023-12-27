package signup

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"
	"unicode"

	"github.com/rogue-syntax/rs-goapiserver/apierrors"
	"github.com/rogue-syntax/rs-goapiserver/apireturn"

	"github.com/rogue-syntax/rs-goapiserver/apireturn/apierrorkeys"
	"github.com/rogue-syntax/rs-goapiserver/authentication"
	"github.com/rogue-syntax/rs-goapiserver/authutil"
	"github.com/rogue-syntax/rs-goapiserver/database"
	"github.com/rogue-syntax/rs-goapiserver/global"
	"github.com/rogue-syntax/rs-goapiserver/mail"
	//"github.com/Jeffail/gabs/v2"
)

type PasswordReset struct {
	Password_reset_id      *int
	Email_address          *string
	Password_reset_token   *string
	Password_reset_expires *int
	User_id                *int
}

type EmailAvailable struct {
	EmailAvailable bool
	EmailSent      bool
	Trace          float32
	ErrorMsg       string
}

type EmailSubmission struct {
	EmailAddress string
}

type EmailVerification struct {
	Email_verification_id *int
	Email_address         *string
	Email_verif_token     *string
	Email_verif_expires   *int
	Password_reset_id     *int
}

type TokenValidation struct {
	Token   string
	NewUser bool
}

type TokenValidationResponse struct {
	IsValid  bool
	PwToken  string
	Trace    float32
	ErrorMsg string
}

type PWSubmission struct {
	NewPw   string
	PwToken string
	NewUser bool
}

type PWValidationResponse struct {
	IsValid  bool
	Trace    int
	PwToken  string
	ErrorMsg string
	PwReqMsg string
}

func pwIsValid(s string) bool {
	var (
		hasMinLen  = false
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)
	if len(s) >= 9 {
		hasMinLen = true
	}
	for _, char := range s {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}
	return hasMinLen && hasUpper && hasLower && hasNumber && hasSpecial
}

func verifyEmail(emailString *string) bool {
	re := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	return re.MatchString((*emailString))
}

func checkEmailUnique(email_value string) (bool, error) {
	var count *int
	err := database.DB.Get(&count, "SELECT COUNT(*) FROM main.email WHERE email_value = ?", email_value)
	if err != nil {
		return false, err
	}

	if *count > 0 {
		return false, nil
	}
	return true, nil
}

func TestPWVerifEP_handler(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	pw := r.FormValue("pw")
	isValid := pwIsValid(pw)
	var isValidStr string
	if isValid {
		isValidStr = "true"

	} else {
		isValidStr = "false"

	}
	fmt.Fprintf(w, isValidStr)
}

func PWVerifEP_handler(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	//input : PWSubmission
	/*output: PWValidationResponse
	error branches for client:
		- PWReqNotFound : request not found or expired
		- PWReqNotMet : pw reqs not met
		- LoginFailed : unable to log user in
	*/
	var pwSubmission PWSubmission
	var pwValidationResponse PWValidationResponse
	pwValidationResponse.IsValid = false
	//decode PWSubmission object from client
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&pwSubmission)
	if err != nil {
		pwValidationResponse.Trace = 0
		pwValidationResponse.ErrorMsg = err.Error()
		pwValidationResponse.PwReqMsg = apierrorkeys.CantDecode
		apireturn.ApiJSONReturn(pwValidationResponse, apierrorkeys.APIReqError, &w)
		return
	}
	//Get current time
	currentTime := time.Now()
	currentTimeUnix := currentTime.Unix()
	//Clear expired PW Verifications
	_, err = database.DB.Exec("call main.clearExpiredPWVerification(?)", currentTimeUnix)
	if err != nil {
		pwValidationResponse.Trace = 1
		pwValidationResponse.ErrorMsg = err.Error()
		pwValidationResponse.PwReqMsg = apierrorkeys.APIReqError
		apireturn.ApiJSONReturn(pwValidationResponse, apierrorkeys.APIReqError, &w)
		return
	}
	//check to see if PW Verificatiopn record with supplied token exists and is not expired
	var passwordReset PasswordReset
	err = database.DB.Get(&passwordReset, "SELECT * FROM main.password_reset WHERE password_reset_token = ? AND password_reset_expires > ?;", pwSubmission.PwToken, currentTimeUnix)
	//err = database.DB.Get(&passwordReset, "SELECT * FROM main.password_reset WHERE password_reset_token = ? ;", pwSubmission.PwToken)
	if err != nil {
		pwValidationResponse.Trace = 2
		pwValidationResponse.ErrorMsg = err.Error()
		pwValidationResponse.PwReqMsg = pwSubmission.PwToken
		apireturn.ApiJSONReturn(pwValidationResponse, apierrorkeys.APIReqError, &w)
		return
	}
	if *passwordReset.Password_reset_token == "" {
		pwValidationResponse.Trace = 3
		pwValidationResponse.ErrorMsg = "Password reset record expired or not found"
		pwValidationResponse.PwReqMsg = apierrorkeys.PWReqNotFound
		apireturn.ApiJSONReturn(pwValidationResponse, apierrorkeys.APIReqError, &w)
		return
	}
	//password req check
	isValid := pwIsValid(pwSubmission.NewPw)
	if isValid == false {
		pwValidationResponse.Trace = 4
		pwValidationResponse.ErrorMsg = "Password does not meet requirements"
		pwValidationResponse.PwReqMsg = apierrorkeys.PWReqNotMet
		apireturn.ApiJSONReturn(pwValidationResponse, apierrorkeys.APIReqError, &w)
		return
	}
	//generate pw hash
	pwHash, err := authutil.GeneratePW(pwSubmission.NewPw)
	if err != nil {
		pwValidationResponse.Trace = 5
		pwValidationResponse.ErrorMsg = err.Error()
		pwValidationResponse.PwReqMsg = apierrorkeys.APIReqError
		apireturn.ApiJSONReturn(pwValidationResponse, apierrorkeys.APIReqError, &w)
		return
	}
	//set pw for user in db
	_, err = database.DB.Exec("INSERT INTO main.user_auth (user_id, user_pw) VALUES (?, ?) ON DUPLICATE KEY UPDATE user_pw = ?;", *passwordReset.User_id, pwHash, pwHash)

	if err != nil {
		pwValidationResponse.Trace = 6
		pwValidationResponse.ErrorMsg = err.Error()
		pwValidationResponse.PwReqMsg = apierrorkeys.APIReqError
		apireturn.ApiJSONReturn(pwValidationResponse, apierrorkeys.APIReqError, &w)
		return
	}
	//pw should be in system and good to go so log user in
	_, err = authentication.HandleAppBrowserSignIn(pwSubmission.NewPw, *passwordReset.Email_address, w, r)
	if err != nil {
		pwValidationResponse.Trace = 7
		pwValidationResponse.ErrorMsg = err.Error()
		pwValidationResponse.PwReqMsg = "" + pwSubmission.NewPw + ", " + *passwordReset.Email_address
		apireturn.ApiJSONReturn(pwValidationResponse, apierrorkeys.APIReqError, &w)
		return
	}
	//return no error response to client
	pwValidationResponse.Trace = 8
	pwValidationResponse.ErrorMsg = apierrorkeys.NOError
	pwValidationResponse.PwReqMsg = apierrorkeys.NOError
	apireturn.ApiJSONReturn(pwValidationResponse, apierrorkeys.NOError, &w)
	return

}

func EmailVerifEP_handler(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	var tokenValidation TokenValidation
	var validationResp TokenValidationResponse
	validationResp.IsValid = false
	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&tokenValidation)
	if err != nil {
		validationResp.Trace = 0
		apireturn.ApiJSONReturn(validationResp, apierrorkeys.APIReqError, &w)
		return
	}

	currentTime := time.Now()
	pwExpTime := time.Now().Add(15 * time.Minute)
	currentTimeUnix := currentTime.Unix()
	pwExpTimeUnix := pwExpTime.Unix()

	_, err = database.DB.Exec("call main.clearExpiredEMailVerification(?)", currentTimeUnix)
	if err != nil {
		validationResp.Trace = 1
		validationResp.ErrorMsg = err.Error()
		apireturn.ApiJSONReturn(validationResp, apierrorkeys.APIReqError, &w)
		return
	}

	var emailVerification []EmailVerification
	err = database.DB.Select(&emailVerification, "SELECT * FROM main.email_verification WHERE email_verif_token = ? && email_verif_expires > ? ", tokenValidation.Token, currentTimeUnix)

	if err != nil {
		validationResp.Trace = 2
		validationResp.ErrorMsg = err.Error()
		apireturn.ApiJSONReturn(validationResp, apierrorkeys.APIReqError, &w)
		return
	}
	if len(emailVerification) == 0 {
		validationResp.Trace = 3
		apireturn.ApiJSONReturn(validationResp, apierrorkeys.APIReqError, &w)
		return
	}

	pwToken, _, err := authutil.MakeAuthToken()
	if err != nil {
		validationResp.Trace = 4
		validationResp.ErrorMsg = err.Error()
		apireturn.ApiJSONReturn(validationResp, apierrorkeys.APIReqError, &w)
		return
	}

	emailVerif := emailVerification[0]
	//createNewUserFromEmail : email:string, pwRequestTokem:string, time exp: int )
	_, err = database.DB.Exec("call main.createNewUserFromEmail(?,?,?)", emailVerif.Email_address, pwToken, pwExpTimeUnix)
	if err != nil {
		validationResp.Trace = 5
		apireturn.ApiJSONReturn(err.Error(), apierrorkeys.APIReqError, &w)
		return
	}

	validationResp.PwToken = pwToken
	validationResp.IsValid = true

	apireturn.ApiJSONReturn(validationResp, apierrorkeys.NOError, &w)

	return

}

func Handler_AppSignUp(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	//TODO: check source is valid client
	var emailSubmission EmailSubmission
	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&emailSubmission)
	if err != nil {
		apierrors.HandleError(err, err.Error(), &apierrors.ReturnError{Msg: apierrorkeys.SystemError, W: &w})
		return
	}

	// verify email syntax and sanitize
	// check email unique
	//checkEmailUnique
	isUnique, err := checkEmailUnique(emailSubmission.EmailAddress)
	if err != nil {
		apierrors.HandleError(err, err.Error(), &apierrors.ReturnError{Msg: apierrorkeys.EmailTaken, W: &w})
		return
	}

	var isAvailable EmailAvailable
	isAvailable.EmailAvailable = true
	isAvailable.EmailSent = false
	isAvailable.Trace = 0
	if isUnique == false {
		isAvailable.EmailAvailable = false
		apireturn.ApiJSONReturn(isAvailable, apierrorkeys.NOError, &w)
		return
	} else {
		// create token
		token, _, err := authutil.MakeAuthToken()
		if err != nil {
			isAvailable.Trace = 1
			apireturn.ApiJSONReturn(isAvailable, apierrorkeys.APIReqError, &w)
			return
		}
		// get expiration time
		expTime := time.Now().Add(15 * time.Minute)
		expTimeUnix := expTime.Unix()
		// log email verif record to db
		_, err = database.DB.Exec("INSERT INTO main.email_verification ( email_address, email_verif_token, email_verif_expires) VALUES ( ?,?,? );", emailSubmission.EmailAddress, token, expTimeUnix)
		if err != nil {
			isAvailable.Trace = 3
			isAvailable.ErrorMsg = err.Error()
			apireturn.ApiJSONReturn(isAvailable, apierrorkeys.APIReqError, &w)
			return
		}
		// craft verification email
		html := `<span>Welcome to KIBANX.</span><br/><span> Please follow <a href="https://` + global.EnvVars.Apiserver + `/set-pw?token=` + token + `&verifyEmail=true&newUser=true"> >this link< </a> to verify your email address and begin your investor onboarding process!</span>`
		emailBody, err := mail.CraftEmail(html)
		if err != nil {
			isAvailable.Trace = 4
			apireturn.ApiJSONReturn(isAvailable, apierrorkeys.APIReqError, &w)
			return
		}
		err = mail.SendMailSingle(emailSubmission.EmailAddress, emailBody, "KIBANX Support", global.EnvVars.SMTPSupportUserName, "KIBANX email verification")
		// send verification email
		if err != nil {
			isAvailable.Trace = 5
			apireturn.ApiJSONReturn(isAvailable, apierrorkeys.APIReqError, &w)
			return
		}

		isAvailable.EmailSent = true
		apireturn.ApiJSONReturn(isAvailable, apierrorkeys.NOError, &w)
	}

}

func Handler_RequestPasswordReset(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	//TODO: check source is valid client
	var emailSubmission EmailSubmission
	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&emailSubmission)
	if err != nil {
		apierrors.HandleError(err, err.Error(), &apierrors.ReturnError{Msg: apierrorkeys.SystemError, W: &w})
		return
	}

	// verify email syntax and sanitize
	// check email unique
	//checkEmailUnique
	isUnique, err := checkEmailUnique(emailSubmission.EmailAddress)
	if err != nil {
		apierrors.HandleError(err, err.Error(), &apierrors.ReturnError{Msg: apierrorkeys.EmailTaken, W: &w})
		return
	}

	var isAvailable EmailAvailable
	isAvailable.EmailAvailable = true
	isAvailable.EmailSent = false
	isAvailable.Trace = 0
	if isUnique == true {
		isAvailable.EmailAvailable = false
		apireturn.ApiJSONReturn(isAvailable, apierrorkeys.NOError, &w)
		return
	} else {
		// create token
		token, _, err := authutil.MakeAuthToken()
		if err != nil {
			isAvailable.Trace = 1
			apireturn.ApiJSONReturn(isAvailable, apierrorkeys.APIReqError, &w)
			return
		}
		// get expiration time
		expTime := time.Now().Add(15 * time.Minute)
		expTimeUnix := expTime.Unix()
		// log email verif record to db
		_, err = database.DB.Exec("INSERT INTO main.email_verification ( email_address, email_verif_token, email_verif_expires) VALUES ( ?,?,? );", emailSubmission.EmailAddress, token, expTimeUnix)
		if err != nil {
			isAvailable.Trace = 3
			isAvailable.ErrorMsg = err.Error()
			apireturn.ApiJSONReturn(isAvailable, apierrorkeys.APIReqError, &w)
			return
		}
		// craft verification email
		html := `<span>Greetings from KIBANX.</span><br/>
		<span>Someone has requested a password reset for the Kibanx account associated with this email.</span><br/>
		<span> Please follow <a href="https://` + global.EnvVars.Apiserver + `/set-pw?token=` + token + `&verifyEmail=true&newUser=false"> >this link< </a> to reset your password!</span>`
		emailBody, err := mail.CraftEmail(html)
		if err != nil {
			isAvailable.Trace = 4
			apireturn.ApiJSONReturn(isAvailable, apierrorkeys.APIReqError, &w)
			return
		}
		err = mail.SendMailSingle(emailSubmission.EmailAddress, emailBody, "KIBANX Support", global.EnvVars.SMTPSupportUserName, "KIBANX email verification")
		// send verification email
		if err != nil {
			isAvailable.Trace = 5
			apireturn.ApiJSONReturn(isAvailable, apierrorkeys.APIReqError, &w)
			return
		}

		isAvailable.EmailSent = true
		apireturn.ApiJSONReturn(isAvailable, apierrorkeys.NOError, &w)
	}

}
