package global

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/pkg/errors"
	"github.com/rogue-syntax/rs-goapiserver/rs_date"
)

const (
	YYYYMMDD = "2006-01-02"
)

type EnvVarsType struct {
	AppRootDir           string
	PublicRootDir        string
	ServiceName          string
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
	SMTPServiceDomain    string
	SMSTestPhone         string
	TwilTestToken        string
	TwilTestAcct         string
	RecaptchaSecret      string
	RecaptchaEP          string
	RecaptchaThreshold   float64
	MinioEndpoint        string
	MinioAccessKey       string
	MinioSecretAccessKey string
	MinioUseSSL          bool
	MinioSSLKey          string
	MinioSSLCert         string
	GGLMapsApiKey        string
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

func GenerateUniqueString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

type MonthRange int

const (
	January MonthRange = iota
	February
	March
	April
	May
	June
	July
	August
	September
	October
	November
	December
)

var MonthNames = []string{
	"January",
	"February",
	"March",
	"April",
	"May",
	"June",
	"July",
	"August",
	"September",
	"October",
	"November",
	"December",
}

type MonthAndDayCount struct {
	Month MonthRange
	Days  int
}

type MonthLookup struct {
	Name  string
	Index int
}

var MonthLookupList = []MonthLookup{
	{Name: "January", Index: 0},
	{Name: "February", Index: 1},
	{Name: "March", Index: 2},
	{Name: "April", Index: 3},
	{Name: "May", Index: 4},
	{Name: "June", Index: 5},
	{Name: "July", Index: 6},
	{Name: "August", Index: 7},
	{Name: "September", Index: 8},
	{Name: "October", Index: 9},
	{Name: "November", Index: 10},
	{Name: "December", Index: 11},
}

var MonthAndDayCountLookup = []MonthAndDayCount{
	{Month: January, Days: 31},
	{Month: February, Days: 28},
	{Month: March, Days: 31},
	{Month: April, Days: 30},
	{Month: May, Days: 31},
	{Month: June, Days: 30},
	{Month: July, Days: 31},
	{Month: August, Days: 31},
	{Month: September, Days: 30},
	{Month: October, Days: 31},
	{Month: November, Days: 30},
	{Month: December, Days: 31},
}

type CountryName string

const (
	UnitedStates  CountryName = "United States"
	Japan         CountryName = "Japan"
	Canada        CountryName = "Canada"
	EuropeanUnion CountryName = "European Union"
	UnitedKingdom CountryName = "United Kingdom"
	Australia     CountryName = "Australia"
	China         CountryName = "China"
	India         CountryName = "India"
	Switzerland   CountryName = "Switzerland"
	SouthAfrica   CountryName = "South Africa"
	Sweden        CountryName = "Sweden"
	Norway        CountryName = "Norway"
	Denmark       CountryName = "Denmark"
	Poland        CountryName = "Poland"
	Hungary       CountryName = "Hungary"
	CzechRepublic CountryName = "Czech Republic"
	Israel        CountryName = "Israel"
	Turkey        CountryName = "Turkey"
	Russia        CountryName = "Russia"
	Brazil        CountryName = "Brazil"
	Mexico        CountryName = "Mexico"
	Colombia      CountryName = "Colombia"
	Argentina     CountryName = "Argentina"
	Chile         CountryName = "Chile"
	Peru          CountryName = "Peru"
	Venezuela     CountryName = "Venezuela"
	Bolivia       CountryName = "Bolivia"
	Haiti         CountryName = "Haiti"
	Nigeria       CountryName = "Nigeria"
	Zambia        CountryName = "Zambia"
	Kenya         CountryName = "Kenya"
	Uganda        CountryName = "Uganda"
	Tanzania      CountryName = "Tanzania"
	Ghana         CountryName = "Ghana"
	Morocco       CountryName = "Morocco"
	Egypt         CountryName = "Egypt"
)

type FiscalYearEndDay struct {
	Country CountryName
	Day     int
}

type FiscalYearEndMonth struct {
	Country CountryName
	Month   MonthRange
}

var FiscalYearEndDays = []FiscalYearEndDay{
	{Country: UnitedStates, Day: 31},
	{Country: Japan, Day: 31},
	{Country: Canada, Day: 31},
	{Country: EuropeanUnion, Day: 31},
	{Country: UnitedKingdom, Day: 31},
	{Country: Australia, Day: 31},
	{Country: China, Day: 31},
	{Country: India, Day: 31},
	{Country: Switzerland, Day: 31},
	{Country: SouthAfrica, Day: 31},
	{Country: Sweden, Day: 31},
	{Country: Norway, Day: 31},
	{Country: Denmark, Day: 31},
	{Country: Poland, Day: 31},
	{Country: Hungary, Day: 31},
	{Country: CzechRepublic, Day: 31},
	{Country: Israel, Day: 31},
	{Country: Turkey, Day: 31},
	{Country: Russia, Day: 31},
	{Country: Brazil, Day: 31},
	{Country: Mexico, Day: 31},
	{Country: Colombia, Day: 31},
	{Country: Argentina, Day: 31},
	{Country: Chile, Day: 31},
	{Country: Peru, Day: 31},
	{Country: Venezuela, Day: 31},
	{Country: Bolivia, Day: 31},
	{Country: Haiti, Day: 31},
	{Country: Nigeria, Day: 31},
	{Country: Zambia, Day: 31},
	{Country: Kenya, Day: 31},
	{Country: Uganda, Day: 31},
	{Country: Tanzania, Day: 31},
	{Country: Ghana, Day: 31},
	{Country: Morocco, Day: 31},
	{Country: Egypt, Day: 31},
}

var FiscalYearEndMonths = []FiscalYearEndMonth{
	{Country: UnitedStates, Month: December},
	{Country: Japan, Month: March},
	{Country: Canada, Month: March},
	{Country: EuropeanUnion, Month: December},
	{Country: UnitedKingdom, Month: March},
	{Country: Australia, Month: June},
	{Country: China, Month: December},
	{Country: India, Month: March},
	{Country: Switzerland, Month: December},
	{Country: SouthAfrica, Month: March},
	{Country: Sweden, Month: December},
	{Country: Norway, Month: December},
	{Country: Denmark, Month: December},
	{Country: Poland, Month: December},
	{Country: Hungary, Month: December},
	{Country: CzechRepublic, Month: December},
	{Country: Israel, Month: December},
	{Country: Turkey, Month: December},
	{Country: Russia, Month: December},
	{Country: Brazil, Month: December},
	{Country: Mexico, Month: December},
	{Country: Colombia, Month: December},
	{Country: Argentina, Month: December},
	{Country: Chile, Month: December},
	{Country: Peru, Month: December},
	{Country: Venezuela, Month: December},
	{Country: Bolivia, Month: December},
	{Country: Haiti, Month: December},
	{Country: Nigeria, Month: December},
	{Country: Zambia, Month: December},
	{Country: Kenya, Month: December},
	{Country: Uganda, Month: December},
	{Country: Tanzania, Month: December},
	{Country: Ghana, Month: December},
	{Country: Morocco, Month: December},
	{Country: Egypt, Month: December},
}

func GetLastDayOfMonth(d rs_date.RSDate) rs_date.RSDate {
	year, month, _ := d.Date()
	days := MonthAndDayCountLookup[month-1].Days

	// Adjust for leap year if February
	if int(month) == int(February)+1 && isLeapYear(year) {
		days++
	}

	return rs_date.NewRSDate(year, time.Month(month), days)
}

func isLeapYear(year int) bool {
	if year%4 == 0 {
		if year%100 == 0 {
			return year%400 == 0
		}
		return true
	}
	return false
}
