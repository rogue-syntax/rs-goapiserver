package apierrorkeys

const (
	//APP
	AppInitErr     = "APP_INIT_ERROR"
	DBInitErr      = "DB_INIT_ERR"
	PanicError     = "PANIC_RECOVERY"
	ServeHttpError = "SERVE_HTTP_ERROR"
	SendMailError  = "SEND_MAIL_ERROR"
	SystemError    = "SYSTEM_ERROR"
	//
	NOError             = "NO_ERROR"
	PWIncorrect         = "PASSWORD_INCORRECT"
	JSONMarshalError    = "JSON_MARSHALL_ERROR"
	JSONDecodeError     = "JSON_DECODE_ERROR"
	AuthorizationError  = "AUTH_ERROR"
	MiddlewareError     = "MIDDLEWARE_ERROR"
	SessionExpired      = "SESSION_EXPIRED"
	ContextError        = "CONTEXT_ERROR"
	APIKeyNotFound      = "API_KEY_NOT_FOUND"
	AuthHeaderNotFound  = "AUTH_HEADER_NOT_FOUND"
	APIReqError         = "API_REQ_ERROR"
	PWReqNotMet         = "PW_REQ"
	PWReqNotFound       = "PW_REQ_NOT_FOUND"
	CantDecode          = "CANT_DECODE"
	LoginFailed         = "LOGIN_FAILED"
	LoggedOut           = "LOGGED_OUT"
	UnautorizedResource = "UNAUTHORIZED RESOURCE"
	DBExecError         = "DB_EXEC_ERROR"
	SMSSendError        = "SMS_SEND_ERROR"
	SMSMsgSent          = "MESSAGE_SENT"
	SMSCodeNotFound     = "SMS_CODE_NOT_FOUND"
	HTTPPostReqError    = "HTTP_POST_REQ_ERROR"
	DBQueryError        = "DB_QUERY_ERROR"
	RowScanError        = "ROW_SCAN_ERROR"
	S3ReadError         = "S3_READ_ERROR"
	S3WriteError        = "S3_WRITE_ERROR"
	WebSocketError      = "WEBSOCKET_ERROR"

	//signuo
	EmailTaken = "EMAIL_TAKEN"
)
