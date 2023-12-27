package global

import (
	"encoding/json"
	"io/ioutil"

	"github.com/pkg/errors"
)

const (
	YYYYMMDD = "2006-01-02"
)

type EnvVarsType struct {
	Apiserver            string
	Dbserver             string
	DbserverPW           string
	DbserverUser         string
	DbserverPort         string
	DbserverDefaultDB    string
	DBTLS                bool
	SSLCliKey            string
	SSLCliCert           string
	SSLCaCert            string
	DevEnv               bool
	TestKey              string
	SMTPEndpoint         string
	SMTPPort             string
	SMTPSupportUserName  string
	SMTPSupportUserPW    string
	SMTPServiceAPIKey    string
	SMSTestPhone         string
	TwilTestToken        string
	TwilTestAcct         string
	RecaptchaSecret      string
	RecaptchaEP          string
	RecaptchaThreshold   float64
	MinioEndpoint        string
	MinioAccessKey       string
	MinioSecretAccessKey string
}

var Reference_YYYY_MM_DD = "2006-01-02"

var EnvVars EnvVarsType

var AuthTimeout int64 = 8

// InitEnvVars
//
//	-Look to command line args or flags for environemnt variables
//	-Look to "/var/www/env/blockEnv.txt" for environment variables
//	-"/var/www/env/blockEnv.txt" should be off limits to dev user but not goapiserver user
func InitEnvVars() error {

	blockEnvTxt, err := ioutil.ReadFile("/var/env/env.json")
	json.Unmarshal(blockEnvTxt, &EnvVars)

	if err != nil {
		errx := errors.Wrap(err, err.Error())
		return errx
	}
	return nil
}
