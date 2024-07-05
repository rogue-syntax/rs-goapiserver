package apierrorkeys

const (
	// APP
	AppInitErr     = "APP_INIT_ERROR"
	AppInitErr_DB  = "APP_INIT_ERROR_DB"
	AppInitErr_ENV = "APP_INIT_ERROR_DB_ENV"
	AppInitErr_S3  = "APP_INIT_ERROR_DB_S3"
	DBInitErr      = "DB_INIT_ERR"
	PanicError     = "PANIC_RECOVERY"
	ServeHttpError = "SERVE_HTTP_ERROR"
	SendMailError  = "SEND_MAIL_ERROR"
	LogGenError    = "LOG_GEN_ERROR"

	// Authentication
	AuthorizationError = "AUTH_ERROR"
	MiddlewareError    = "MIDDLEWARE_ERROR"
	SessionExpired     = "SESSION_EXPIRED"
	APIKeyNotFound     = "API_KEY_NOT_FOUND"
	AuthHeaderNotFound = "AUTH_HEADER_NOT_FOUND"

	//Account
	AccountError                  = "ACCOUNT_ERROR"
	NonexistentAccount            = "NONEXISTENT_ACCOUNT"
	DefunctCompanyMemeberships    = "DEFUNCT_COMPANY_MEMBERSHIPS"
	NoCompanyMemeberships         = "NO_COMPANY_MEMBERSHIPS"
	COMPANY_AUTHNTICATED_MISMATCH = "COMPANY_AUTHNTICATED_MISMATCH"

	// Password
	PWIncorrect   = "PASSWORD_INCORRECT"
	PWReqNotMet   = "PW_REQ"
	PWReqNotFound = "PW_REQ_NOT_FOUND"
	CantDecode    = "CANT_DECODE"
	LoginFailed   = "LOGIN_FAILED"
	LoggedOut     = "LOGGED_OUT"
	SignupError   = "SIGNUP_ERROR"
	EmailTaken    = "EMAIL_TAKEN"

	// Data
	JSONMarshalError    = "JSON_MARSHALL_ERROR"
	JSONDecodeError     = "JSON_DECODE_ERROR"
	ContextError        = "CONTEXT_ERROR"
	DataConversionError = "DATA_CONVERSION_ERROR"

	// API Requests
	APIReqError = "API_REQ_ERROR"

	// Database
	DBExecError  = "DB_EXEC_ERROR"
	DBQueryError = "DB_QUERY_ERROR"
	RowScanError = "ROW_SCAN_ERROR"

	// File Operations
	FileUploadError      = "FileUploadError"
	FileEmpty            = "FILE_EMPTY"
	MapKeyNotFound       = "MAP_KEY_NOT_FOUND"
	UnauthorizedFileType = "UNAUTHORIZED_FILE_TYPE"
	MismatchedFileType   = "MISMATCHED_FILE_TYPE"

	// SMS
	SMSSendError    = "SMS_SEND_ERROR"
	SMSMsgSent      = "MESSAGE_SENT"
	SMSCodeNotFound = "SMS_CODE_NOT_FOUND"

	// HTTP
	HTTPPostReqError = "HTTP_POST_REQ_ERROR"

	// AWS S3
	S3ReadError  = "S3_READ_ERROR"
	S3WriteError = "S3_WRITE_ERROR"

	// WebSockets
	WebSocketError = "WEBSOCKET_ERROR"

	// Utility
	UtilityProcessError = "UTILITY_PROCESS_ERROR"

	// General
	NOError     = "NO_ERROR"
	SystemError = "SYSTEM_ERROR"

	//Input Integrity Errors
	InvalidAPIInput = "INVALID_INPUT"

	//event system error
	EventError = "EVENT_ERROR"

	FormFieldValidationError = "FORM_FIELD_VALIDATION_ERROR"

	FormFieldUserError = "FORM_FIELD_USER_ERROR"

	NullDataError = "NULL_DATA_ERROR"

	InvalidType = "INVALID_TYPE"

	InvalidCast = "INVALID_CAST"

	AutoIdOnNewRecord = "AUTO_ID_ON_NEW_RECORD"

	RequiredDataMissing = "REQUIRED_DATA_MISSING"

	UDFMarshalError = "UDF_MARSHAL_ERROR"

	UDFBrokerError = "UDF_BROKER_ERROR"

	UDFADDError = "UDF_ADD_ERROR"

	RequestError = "REQUEST_ERROR"
)
