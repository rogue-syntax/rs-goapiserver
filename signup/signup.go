package signup

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"
	"unicode"

	"github.com/nbutton23/zxcvbn-go"
	"github.com/pkg/errors"
	"github.com/rogue-syntax/rs-goapiserver/apierrors"
	"github.com/rogue-syntax/rs-goapiserver/apimaster"
	"github.com/rogue-syntax/rs-goapiserver/apireturn"
	"github.com/rogue-syntax/rs-goapiserver/form_verification"
	"github.com/rogue-syntax/rs-goapiserver/rs_zxcvbn"
	"github.com/rogue-syntax/rs-goapiserver/sql_tools"

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

func CheckEmailUnique(email_value string) (bool, error) {
	var count *int
	err := database.DB.Get(&count, "SELECT COUNT(*) FROM user_base WHERE user_email = ?", email_value)
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

// "/v1/signup/pw-verif-ep" WVerifEP_handler PWVerifEP_handler_ApiReq
var PWVerifEP_handler_ApiReq apimaster.ApiReqDef = apimaster.ApiReqDef{
	API:           "/v1/signup/pwVerifEp",
	Method:        apimaster.POSTREQ,
	Desc:          "Password Verification Endpoint",
	Input:         apimaster.MakeStructDescriptorMap(new(PWSubmission)),
	OutputData:    apimaster.MakeStructDescriptorMap(new(PWValidationResponse)),
	OutputWrapper: apimaster.MakeStructDescriptorMap(new(apireturn.JsonReturn)),
}

func PWVerifEP_handler(w http.ResponseWriter, r *http.Request, ctx context.Context) {

	var pwSubmission PWSubmission
	var pwValidationResponse PWValidationResponse
	pwValidationResponse.IsValid = false
	//decode PWSubmission object from client
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&pwSubmission)
	if err != nil {
		apierrors.HandleError(r, errors.WithStack(err), apierrorkeys.DBQueryError, &apierrors.ReturnError{Msg: apierrorkeys.DBQueryError, W: &w})
		return
	}
	//Get current time
	currentTime := time.Now()
	currentTimeUnix := currentTime.Unix()
	//Clear expired PW Verifications
	/*
		_, err = database.DB.Exec("call clearExpiredPWVerification(?)", currentTimeUnix)
		if err != nil {
			pwValidationResponse.Trace = 1
			pwValidationResponse.ErrorMsg = err.Error()
			pwValidationResponse.PwReqMsg = apierrorkeys.APIReqError
			apireturn.ApiJSONReturn(pwValidationResponse, apierrorkeys.APIReqError, &w)
			return
		}
	*/
	//check to see if PW Verificatiopn record with supplied token exists and is not expired
	var passwordReset PasswordReset
	err = database.DB.Get(&passwordReset, "SELECT * FROM password_reset WHERE password_reset_token = ? AND password_reset_expires > ?;", pwSubmission.PwToken, currentTimeUnix)
	//err = database.DB.Get(&passwordReset, "SELECT * FROM main.password_reset WHERE password_reset_token = ? ;", pwSubmission.PwToken)
	if err != nil {
		apierrors.HandleError(r, errors.WithStack(err), apierrorkeys.DBQueryError, &apierrors.ReturnError{Msg: apierrorkeys.DBQueryError, W: &w})
		return
	}
	if *passwordReset.Password_reset_token == "" {
		apierrors.HandleError(r, errors.New(apierrorkeys.PWReqNotFound), apierrorkeys.PWReqNotFound, &apierrors.ReturnError{Msg: apierrorkeys.PWReqNotFound, W: &w})
		return
	}

	if pwSubmission.NewUser == false && pwSubmission.NewPw == "" {
		//return sucess
		pwValidationResponse.IsValid = true
		pwValidationResponse.PwToken = *passwordReset.Password_reset_token
		pwValidationResponse.ErrorMsg = ""
		apireturn.ApiJSONReturn(pwValidationResponse, apierrorkeys.NOError, &w)
		return
	}

	//password req check
	//isValid := pwIsValid(pwSubmission.NewPw)\
	trimmedPW, err := form_verification.TrimAndVerifyString(pwSubmission.NewPw, form_verification.NO_RULESET)
	if err != nil {
		apierrors.HandleError(r, errors.WithStack(err), apierrorkeys.PWReqNotMet, &apierrors.ReturnError{Msg: apierrorkeys.PWReqNotMet, W: &w})
		return
	}
	isNotSafe := sql_tools.CheckForSQLInjection(trimmedPW)
	if isNotSafe {
		apierrors.HandleError(r, errors.New(apierrorkeys.PWReqNotMet), apierrors.DO_NOT_LOG_ERROR, &apierrors.ReturnError{Msg: apierrorkeys.PWReqNotMet, W: &w})
		return
	}
	pwStrength := zxcvbn.PasswordStrength(trimmedPW, []string{})
	if pwStrength.Score < rs_zxcvbn.MIN_PW_STRENGTH_SCORE {
		apierrors.HandleError(r, errors.New(apierrorkeys.PWReqNotMet), apierrors.DO_NOT_LOG_ERROR, &apierrors.ReturnError{Msg: apierrorkeys.PWReqNotMet, W: &w})
		return
	}

	//generate pw hash
	pwHash, err := authutil.GeneratePW(trimmedPW)
	if err != nil {
		apierrors.HandleError(r, errors.WithStack(err), apierrorkeys.APIReqError, &apierrors.ReturnError{Msg: apierrorkeys.APIReqError, W: &w})
		return
	}
	//set pw for user in db
	_, err = database.DB.Exec("INSERT INTO user_auth (user_id, user_pw) VALUES (?, ?) ON DUPLICATE KEY UPDATE user_pw = ?;", *passwordReset.User_id, pwHash, pwHash)

	if err != nil {
		apierrors.HandleError(r, errors.WithStack(err), apierrorkeys.APIReqError, &apierrors.ReturnError{Msg: apierrorkeys.APIReqError, W: &w})
		return
	}
	//pw should be in system and good to go so log user in
	_, err = authentication.HandleAppBrowserSignIn(trimmedPW, *passwordReset.Email_address, w, r)
	if err != nil {
		apierrors.HandleError(r, errors.WithStack(err), apierrorkeys.LoginFailed, &apierrors.ReturnError{Msg: apierrorkeys.LoginFailed, W: &w})
		return
	}
	//return no error response to client
	pwValidationResponse.IsValid = true
	pwValidationResponse.PwToken = *passwordReset.Password_reset_token
	pwValidationResponse.ErrorMsg = ""
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
		apierrors.HandleError(r, errors.Wrap(err, "failed to decode token validation"), apierrorkeys.CantDecode, &apierrors.ReturnError{Msg: apierrorkeys.CantDecode, W: &w})
		return
	}

	currentTime := time.Now()
	pwExpTime := time.Now().Add(15 * time.Minute)
	currentTimeUnix := currentTime.Unix()
	pwExpTimeUnix := pwExpTime.Unix()

	_, err = database.DB.Exec("call clearExpiredEMailVerification(?)", currentTimeUnix)
	if err != nil {
		apierrors.HandleError(r, errors.Wrap(err, "failed to clear expired email verifications"), apierrorkeys.DBQueryError, &apierrors.ReturnError{Msg: apierrorkeys.DBQueryError, W: &w})
		return
	}

	var emailVerification []EmailVerification
	err = database.DB.Select(&emailVerification, "SELECT * FROM email_verification WHERE email_verif_token = ? && email_verif_expires > ? ", tokenValidation.Token, currentTimeUnix)

	if err != nil {
		apierrors.HandleError(r, errors.Wrap(err, "failed to query email verification token"), apierrorkeys.DBQueryError, &apierrors.ReturnError{Msg: apierrorkeys.DBQueryError, W: &w})
		return
	}
	if len(emailVerification) == 0 {
		apierrors.HandleError(r, errors.New("email verification token not found or expired"), apierrors.DO_NOT_LOG_ERROR, &apierrors.ReturnError{Msg: apierrorkeys.APIReqError, W: &w})
		return
	}

	pwToken, _, err := authutil.MakeAuthToken()
	if err != nil {
		apierrors.HandleError(r, errors.Wrap(err, "failed to generate authentication token"), apierrorkeys.APIReqError, &apierrors.ReturnError{Msg: apierrorkeys.APIReqError, W: &w})
		return
	}

	emailVerif := emailVerification[0]
	//createNewUserFromEmail : email:string, pwRequestTokem:string, time exp: int )
	_, err = database.DB.Exec("call createNewUserFromEmail(?,?,?)", emailVerif.Email_address, pwToken, pwExpTimeUnix)
	if err != nil {
		apierrors.HandleError(r, errors.Wrap(err, "failed to create new user from email"), apierrorkeys.DBQueryError, &apierrors.ReturnError{Msg: apierrorkeys.DBQueryError, W: &w})
		return
	}

	validationResp.PwToken = pwToken
	validationResp.IsValid = true

	apireturn.ApiJSONReturn(validationResp, apierrorkeys.NOError, &w)

	return

}

// NEEDS TRANSLATION
func Handler_AppSignUp(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	//TODO: check source is valid client
	var emailSubmission EmailSubmission
	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&emailSubmission)
	if err != nil {
		apierrors.HandleError(r, errors.Wrap(err, "failed to decode email submission"), apierrorkeys.CantDecode, &apierrors.ReturnError{Msg: apierrorkeys.CantDecode, W: &w})
		return
	}

	// verify email syntax and sanitize
	// check email unique
	//checkEmailUnique
	isUnique, err := CheckEmailUnique(emailSubmission.EmailAddress)
	if err != nil {
		apierrors.HandleError(r, errors.Wrap(err, "failed to check email uniqueness"), apierrorkeys.DBQueryError, &apierrors.ReturnError{Msg: apierrorkeys.DBQueryError, W: &w})
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
		_, err = database.DB.Exec("INSERT INTO email_verification ( email_address, email_verif_token, email_verif_expires) VALUES ( ?,?,? );", emailSubmission.EmailAddress, token, expTimeUnix)
		if err != nil {
			apierrors.HandleError(r, errors.Wrap(err, "failed to insert email verification record"), apierrorkeys.DBQueryError, &apierrors.ReturnError{Msg: apierrorkeys.DBQueryError, W: &w})
			return
		}

		// craft verification email
		html := `<span>Welcome to ` + global.EnvVars.ServiceName + `.</span><br/><span> Please follow <a href="https://` + global.EnvVars.Apiserver + `/set-pw?token=` + token + `&verifyEmail=true&newUser=true"> >this link< </a> to verify your email address and begin your investor onboarding process!</span>`
		emailBody, err := mail.CraftEmail(html)
		if err != nil {
			apierrors.HandleError(r, errors.Wrap(err, "failed to craft verification email"), apierrorkeys.APIReqError, &apierrors.ReturnError{Msg: apierrorkeys.APIReqError, W: &w})
			return
		}

		err = mail.SendMailSingle(emailSubmission.EmailAddress, emailBody, global.EnvVars.ServiceName+" Support", global.EnvVars.SMTPSupportUserName, global.EnvVars.ServiceName+" email verification")
		// send verification email
		if err != nil {
			apierrors.HandleError(r, errors.Wrap(err, "failed to send verification email"), apierrorkeys.APIReqError, &apierrors.ReturnError{Msg: apierrorkeys.APIReqError, W: &w})
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
		apierrors.HandleError(r, errors.Wrap(err, "failed to decode email submission"), apierrorkeys.CantDecode, &apierrors.ReturnError{Msg: apierrorkeys.CantDecode, W: &w})
		return
	}

	// verify email syntax and sanitize
	// check email unique
	//checkEmailUnique
	isUnique, err := CheckEmailUnique(emailSubmission.EmailAddress)
	if err != nil {
		apierrors.HandleError(r, errors.Wrap(err, "failed to check email uniqueness"), apierrorkeys.DBQueryError, &apierrors.ReturnError{Msg: apierrorkeys.DBQueryError, W: &w})
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
			apierrors.HandleError(r, errors.Wrap(err, "failed to generate authentication token"), apierrorkeys.APIReqError, &apierrors.ReturnError{Msg: apierrorkeys.APIReqError, W: &w})
			return
		}
		// get expiration time
		expTime := time.Now().Add(15 * time.Minute)
		expTimeUnix := expTime.Unix()
		// log email verif record to db
		_, err = database.DB.Exec("INSERT INTO main.email_verification ( email_address, email_verif_token, email_verif_expires) VALUES ( ?,?,? );", emailSubmission.EmailAddress, token, expTimeUnix)
		if err != nil {
			apierrors.HandleError(r, errors.Wrap(err, "failed to insert email verification record"), apierrorkeys.DBQueryError, &apierrors.ReturnError{Msg: apierrorkeys.DBQueryError, W: &w})
			return
		}
		// craft verification email
		html := `<span>Greetings from KIBANX.</span><br/>
		<span>Someone has requested a password reset for the Kibanx account associated with this email.</span><br/>
		<span> Please follow <a href="https://` + global.EnvVars.Apiserver + `/set-pw?token=` + token + `&verifyEmail=true&newUser=false"> >this link< </a> to reset your password!</span>`
		emailBody, err := mail.CraftEmail(html)
		if err != nil {
			apierrors.HandleError(r, errors.Wrap(err, "failed to craft password reset email"), apierrorkeys.APIReqError, &apierrors.ReturnError{Msg: apierrorkeys.APIReqError, W: &w})
			return
		}
		err = mail.SendMailSingle(emailSubmission.EmailAddress, emailBody, "KIBANX Support", global.EnvVars.SMTPSupportUserName, "KIBANX email verification")
		// send verification email
		if err != nil {
			apierrors.HandleError(r, errors.Wrap(err, "failed to send password reset email"), apierrorkeys.APIReqError, &apierrors.ReturnError{Msg: apierrorkeys.APIReqError, W: &w})
			return
		}

		isAvailable.EmailSent = true
		apireturn.ApiJSONReturn(isAvailable, apierrorkeys.NOError, &w)
	}

}
