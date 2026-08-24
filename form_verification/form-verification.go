package form_verification

import (
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/rogue-syntax/rs-goapiserver/apierrors"
	"github.com/rogue-syntax/rs-goapiserver/reflection_tools"
	"github.com/rogue-syntax/rs-goapiserver/sql_tools"

	"github.com/rogue-syntax/rs-goapiserver/apireturn/apierrorkeys"
	"github.com/rogue-syntax/rs-goapiserver/global/util"
)

type RegExRuleset string

const (
	NO_RULESET RegExRuleset = ""
	//"Only letters, numbers, spaces, and special characters _ - @ ( ) up to ten characters."
	INPUT_STRING_RULESET_1 RegExRuleset = `^[a-zA-Z0-9_\\@() -.,]{1,15}$`
	//"Only letters, numbers, spaces, and special characters _ - @ ( ) up to one hundred characters."
	INPUT_STRING_RULESET_2 RegExRuleset = "^[a-zA-Z0-9_\\@() -.,/]{1,100}$"
	//"Only letters, numbers, spaces, and special characters _ - @ ( ) up to 255 characters."
	//INPUT_STRING_RULESET_CHARBLOCK RegExRuleset = "^[a-zA-Z0-9_\\@() -]{1,255}$"
	INPUT_STRING_RULESET_CHARBLOCK RegExRuleset = "^[a-zA-Z0-9_\\@()\\-'.,%\\n\\r ]{0,255}$"
	//"Only letters, numbers, underscore and dash."
	INPUT_STRING_RULESET_MAP_KEYS RegExRuleset = "^[a-zA-Z0-9_\\@()\\-\\$% ]{1,100}$"

	INPUT_STRING_RULESET_CODE_PREFIX RegExRuleset = "^[a-zA-Z0-9_]{1,20}$"

	//need a rule set for legitimate excel cell names
	INPUT_STRING_RULESET_EXCEL_CELL RegExRuleset = "^[A-Z]{1,2}[0-9]{1,2}$"

	//need a rulset for a valid excel named
	INPUT_STRING_RULESET_EXCEL_RANGE RegExRuleset = "^[a-zA-Z_][a-zA-Z0-9_]*$"

	//need a rulset for a valid excel range name
	INPUT_STRING_RULESET_CELL_NAME RegExRuleset = "^[A-Z]{1,2}[0-9]{1,7}$"

	//need a rulset for a valid excel range name
	INPUT_STRING_RULESET_RANGE_NAME RegExRuleset = "^[A-Z]{1,2}[0-9]{1,7}$"

	INPUT_STRING_RULESET_EXCEL_ANY RegExRuleset = `^([A-Z]{1,2}[0-9]{1,7}|[A-Z]{1,2}[0-9]{1,7}:[A-Z]{1,2}[0-9]{1,7}|[a-zA-Z_][a-zA-Z0-9_]*)$`
)

var InputRulesetDescriptions = map[RegExRuleset]string{
	NO_RULESET:                       "No ruleset defined",
	INPUT_STRING_RULESET_1:           "Only letters, numbers, spaces, and special characters: '_ - @ ( ) . ,', up to ten characters.",
	INPUT_STRING_RULESET_2:           "Only letters, numbers, spaces, and special characters: '_ - @ ( ) . , ', / up to one hundred characters.",
	INPUT_STRING_RULESET_CHARBLOCK:   "Only letters, numbers, spaces, and special characters: '_ - @ ( ) % . ' , ', up to two hundred fifty-five characters.",
	INPUT_STRING_RULESET_MAP_KEYS:    "Only letters, numbers, spaces, and special characters: '_ - @ ( ) % $', up to one hundred characters.",
	INPUT_STRING_RULESET_CODE_PREFIX: "Only letters and numbers, and underscore - no hyphen, spaces, up to twenty characters.",
	INPUT_STRING_RULESET_EXCEL_CELL:  "Must be a valid excel cell name.",
	INPUT_STRING_RULESET_EXCEL_RANGE: "Must be a valid excel range name.",
}

func TrimAndVerifyString(input string, ruleset RegExRuleset) (string, error) {
	trimmed := strings.TrimSpace(input)
	if !VerifyInputString(trimmed, ruleset) { //if this doesnt comply
		return "", errors.New(InputRulesetDescriptions[ruleset])
	}
	return trimmed, nil
}

func IsErrorRegexRuleset(err error) bool {
	msgStr := RegExRuleset(err.Error())
	if _, ok := InputRulesetDescriptions[msgStr]; !ok {
		return true
	} else {
		return false
	}
}

/*
Verify that input string matches a given RegExRuleset.
*/
func VerifyInputString(input string, ruleset RegExRuleset) bool {
	regex := regexp.MustCompile(string(ruleset))
	return regex.MatchString(input)
}

const (
	//NEVER SEND IGNORE ERROR TO CLIENT
	IGNORE_ERROR                     = "No error"
	NO_MATCHING_ERROR                = "No matching error found"
	ERROR_PROCESSING_REQUEST         = "Error processing request"
	DATA_OK                          = "Data is ok"
	NO_ID                            = "Id is required"
	NO_ID_IN_DB                      = "Id is not a valid fund_id in the database"
	NO_ID_COMPANY                    = "Id does not belong to the current company"
	NO_NAME                          = "Name is required"
	NO_SHORT_NAME                    = "Short name is required"
	NAME_BAD_REQ                     = "Same does not meet requirements"
	SHORT_NAME_BAD_REQ               = "Short name does not meet requirement"
	NAME_TAKEN                       = "Name is already taken"
	SHORT_NAME_TAKEN                 = "Short name is already taken"
	CODE_TAKEN                       = "Code is already taken"
	DUPLICATE_ADD                    = "New record should not have an ID. Possible duplicate?"
	NO_CODE                          = "Code is required"
	NO_CODE_PREFIX                   = "Code prefix is required"
	MISSING_REQUIRED_REFERENCE_FEILD = "Missing required reference field"
	FIELD_AND_ID_MISMATCH            = "Unique field and id do not match."
	UNIQUE_TYPE_PER_YEAR             = "Only one unique type of this kind per year"
	ALIAS_TAKEN                      = "Alias is already taken"

	//
	REFERENCE_EXISTS = "A reference to this record exists in another table"

	CELL_NAME = "Must be a valid cell name"

	CELL_ORDER = "End cell must be greater than start cell"

	RANGE_NAME = "Must be a valid range name"

	MAP_KEY = "Must be a valid map key"

	SHEET_NAME = "Must be a valid sheet name"

	DATA_MISSING = "Nil value, data is missing"

	START_CELL_MISSING = "Start cell is required"

	END_CELL_LESS_THAN_START_CELL = "End cell must be greater than start cell"

	NO_EXCEL_MAP_RANGE_DEFINED = "At least a start cell or range name is required"
)

type SimpleStringFormFieldVerif struct {
	Value       string
	RuleSet     RegExRuleset
	UniqueCheck UniqueStringCheckStruct
}

type UniqueStringCheckStruct struct {
	TableName        string
	FieldName        string
	QueryComparisons []sql_tools.SimpleQueryComparison
}

func DoSimpleStringFieldVerif(params SimpleStringFormFieldVerif) (verifyResponse string, err error) {

	//check for valid input
	if !VerifyInputString(params.Value, params.RuleSet) {
		msg := InputRulesetDescriptions[params.RuleSet]
		return msg, errors.New(msg)
	}

	taken, err := sql_tools.IsStringTaken(params.UniqueCheck.TableName, params.Value, params.UniqueCheck.FieldName, params.UniqueCheck.QueryComparisons)
	if err != nil {
		return apierrorkeys.DBQueryError, err
	}
	if taken {
		return NAME_TAKEN, errors.New(NAME_TAKEN)
	}

	return string(NO_MATCHING_ERROR), nil
}

func DoSimpleRefCheck(params SimpleStringFormFieldVerif) (verifyResponse string, err error) {

	count, err := sql_tools.SimpleRefCount(params.UniqueCheck.TableName, params.Value, params.UniqueCheck.FieldName, params.UniqueCheck.QueryComparisons)
	if err != nil {
		return apierrorkeys.DBQueryError, err
	}
	if count > 0 {
		return REFERENCE_EXISTS, errors.New(REFERENCE_EXISTS)
	}

	return string(NO_MATCHING_ERROR), nil
}

// use if you want to verify one thing just one way i.e. if name shows up twice when field one is x, no other scenarios

type BasicDBStructVerifObject[StructType any] struct {
	TableDesinationName        string
	StringFieldsForExistCheck  []StringFieldAndError
	AnyFieldsForExistCheck     []AnyFieldAndError
	StringFieldsForUniqueCheck []StringFieldAndError
	StringFieldsForRuleCheck   []StringFieldAndRuleset
	Data                       *StructType
	UniqueIds                  []sql_tools.UniqueId
}

// use if you want to verify one thing multiple ways i.e. if name shows up twice when field one is x but also if field two is y
type BasicDBStructVerifObjectMulti[StructType any] struct {
	TableDesinationName        string
	StringFieldsForExistCheck  []StringFieldAndError
	AnyFieldsForExistCheck     []AnyFieldAndError
	StringFieldsForUniqueCheck []StringFieldAndError
	StringFieldsForRuleCheck   []StringFieldAndRuleset
	Data                       *StructType
	UniqueIds                  [][]sql_tools.UniqueId
}

// going to return DATA_OK if everything is ok
// going to return err != nil if there is a programmatic error
// going to return a message from form_verification consts if there is a validation error
func VerifyBasicDBUpdate[StructType any](basicVerifObj *BasicDBStructVerifObject[StructType]) (string, error) {

	//HANDLE NO DATA INSTANCE FOR STRINGS
	for _, field := range basicVerifObj.StringFieldsForExistCheck {
		//not nil
		if field.Value == nil {
			return field.Err, nil
		}
		//not empty

		if strings.TrimSpace(*field.Value) == "" {
			return field.Err, nil
		}

	}

	for _, field := range basicVerifObj.AnyFieldsForExistCheck {

		val := reflect.ValueOf(field.Value)
		kind := val.Kind()
		//not nil
		if kind == reflect.Ptr && val.IsNil() {
			return field.Err, nil
		}
		//not empty string
		if val.Elem().Kind() == reflect.String {
			theString := val.Elem().Interface().(string)
			if theString == "" {
				//if strings.TrimSpace(theString) == "" {
				return field.Err, nil
			}
		}
		//not empty anything else primitive
		if reflect.DeepEqual(field.Value, reflect.Zero(reflect.TypeOf(field.Value)).Interface()) {
			return field.Err, nil
		}
	}

	//HANDLE RULES MATCHING FOR STRINGS
	for _, field := range basicVerifObj.StringFieldsForRuleCheck {
		if !VerifyInputString(*field.Value, field.RuleSet) {
			msg := InputRulesetDescriptions[field.RuleSet]
			return msg, nil
		}
	}

	//HANDLE NOT UNIQUE SCENARIO FOR STRINGS
	uniqueCheckConstraints := []sql_tools.SimpleQueryComparison{}
	//first generate unique check constraints
	for _, uniqueId := range basicVerifObj.UniqueIds {
		//CEHCK THAT UNIQUE ID IS NOT NIL
		if uniqueId.Unique_id_value == nil {
			return NO_ID, nil
		}
		//CHECK THAT UNIQUE ID IS NOT EMPTY
		unique_id_string, err := util.DistilInterfaceString(uniqueId.Unique_id_value)
		if err != nil {
			return NO_ID, err
		}
		if unique_id_string == "" {
			return NO_ID, nil
		}
		uniqueCheckConstraints = append(uniqueCheckConstraints, sql_tools.SimpleQueryComparison{AndOr: sql_tools.And, Field: uniqueId.Unique_id_field_name, Value: unique_id_string, Comparator: uniqueId.Compartitor})
	}
	//unique check constraints are generated, now check the string fields
	for _, field := range basicVerifObj.StringFieldsForUniqueCheck {
		//ASSUMES ALL TABLE FIELD NAMES MATCH ALL STRUCT FIELD NAMES
		var fieldName string
		if field.FieldName != nil {
			fieldName = *field.FieldName
		} else {
			fieldName = reflection_tools.GetFieldNameByValue(basicVerifObj.Data, field.Value)
		}

		// Check if the field value is taken in the database

		taken, err := sql_tools.IsStringTaken(basicVerifObj.TableDesinationName, *field.Value, fieldName, uniqueCheckConstraints)
		if err != nil {
			//return field.Err, errors.New(field.Err)
			return field.Err, err
		}
		if taken {
			return field.Err, nil
		}

	}
	//now we know,
	//strings are unique and not nil
	//not empty
	//match requirements
	//are unique in the database per SimpleQueryComparison of unique ids
	return DATA_OK, nil
}

func VerifyBasicDBUpdateMulti[StructType any](basicVerifObj *BasicDBStructVerifObjectMulti[StructType]) (string, error) {

	//HANDLE NO DATA INSTANCE FOR STRINGS
	for _, field := range basicVerifObj.StringFieldsForExistCheck {
		//not nil
		if field.Value == nil {
			return field.Err, nil
		}
		//not empty
		if *field.Value == "" {
			return field.Err, nil
		}
	}

	for _, field := range basicVerifObj.AnyFieldsForExistCheck {

		val := reflect.ValueOf(field.Value)
		kind := val.Kind()
		//not nil
		if kind == reflect.Ptr && val.IsNil() {
			return field.Err, nil
		}
		//not empty string
		if val.Elem().Kind() == reflect.String {
			theString := val.Elem().Interface().(string)
			if theString == "" {
				//if strings.TrimSpace(theString) == "" {
				return field.Err, nil
			}
		}
		//not empty anything else primitive
		if reflect.DeepEqual(field.Value, reflect.Zero(reflect.TypeOf(field.Value)).Interface()) {
			return field.Err, nil
		}
	}

	//HANDLE RULES MATCHING FOR STRINGS
	for _, field := range basicVerifObj.StringFieldsForRuleCheck {
		if !VerifyInputString(*field.Value, field.RuleSet) {
			msg := InputRulesetDescriptions[field.RuleSet]
			return msg, nil
		}
	}

	//HANDLE NOT UNIQUE SCENARIO FOR STRINGS
	uniqueCheckConstraints := [][]sql_tools.SimpleQueryComparison{}
	//first generate unique check constraints
	for _, uniqueIds := range basicVerifObj.UniqueIds {

		uniqueCheckConstraint := []sql_tools.SimpleQueryComparison{}

		for _, uniqueId := range uniqueIds {
			//CEHCK THAT UNIQUE ID IS NOT NIL
			if uniqueId.Unique_id_value == nil {
				return NO_ID, nil
			}
			//CHECK THAT UNIQUE ID IS NOT EMPTY
			unique_id_string, err := util.DistilInterfaceString(uniqueId.Unique_id_value)
			if err != nil {
				return NO_ID, err
			}
			if unique_id_string == "" {
				return NO_ID, nil
			}
			uniqueCheckConstraint = append(uniqueCheckConstraint, sql_tools.SimpleQueryComparison{AndOr: sql_tools.And, Field: uniqueId.Unique_id_field_name, Value: unique_id_string, Comparator: uniqueId.Compartitor})

		}
		uniqueCheckConstraints = append(uniqueCheckConstraints, uniqueCheckConstraint)
	}
	//unique check constraints are generated, now check the string fields
	for i, field := range basicVerifObj.StringFieldsForUniqueCheck {
		//ASSUMES ALL TABLE FIELD NAMES MATCH ALL STRUCT FIELD NAMES
		var fieldName string
		if field.FieldName != nil {
			fieldName = *field.FieldName
		} else {
			fieldName = reflection_tools.GetFieldNameByValue(basicVerifObj.Data, field.Value)
		}

		// Check if the field value is taken in the database

		taken, err := sql_tools.IsStringTaken(basicVerifObj.TableDesinationName, *field.Value, fieldName, uniqueCheckConstraints[i])
		if err != nil {
			//return field.Err, errors.New(field.Err)
			return field.Err, err
		}
		if taken {
			return field.Err, nil
		}

	}
	//now we know,
	//strings are unique and not nil
	//not empty
	//match requirements
	//are unique in the database per SimpleQueryComparison of unique ids
	return DATA_OK, nil
}

type ExampleType struct {
	Investment_id         int
	Co_id                 int
	Investment_name       *string
	Investment_short_name *string
}

func VerifyBasicDBUpdate_Example() (string, error) {
	var basicVerifObj BasicDBStructVerifObject[ExampleType]
	var investment ExampleType

	basicVerifObj.StringFieldsForExistCheck = []StringFieldAndError{{Value: investment.Investment_name, Err: NO_NAME}, {Value: investment.Investment_short_name, Err: NO_SHORT_NAME}}
	basicVerifObj.StringFieldsForUniqueCheck = []StringFieldAndError{{Value: investment.Investment_name, Err: SHORT_NAME_TAKEN}, {Value: investment.Investment_short_name, Err: SHORT_NAME_TAKEN}}
	basicVerifObj.StringFieldsForRuleCheck = []StringFieldAndRuleset{{Value: investment.Investment_name, RuleSet: INPUT_STRING_RULESET_2}, {Value: investment.Investment_short_name, RuleSet: INPUT_STRING_RULESET_1}}
	basicVerifObj.Data = &investment
	basicVerifObj.TableDesinationName = "investment"
	basicVerifObj.UniqueIds = []sql_tools.UniqueId{
		{
			Unique_id_value:      investment.Investment_id,
			Unique_id_field_name: "Investment_id",
			AndOr:                sql_tools.And,
			Compartitor:          sql_tools.Equal,
		},
		{
			Unique_id_value:      investment.Co_id,
			Unique_id_field_name: "Co_id",
			AndOr:                sql_tools.And,
			Compartitor:          sql_tools.Equal,
		},
	}

	msg, err := VerifyBasicDBUpdate[ExampleType](&basicVerifObj)
	return msg, err
}

// remember that FiledName must be supplied if struct field is not a pointer
type StringFieldAndError struct {
	Value     *string
	Err       string
	FieldName *string
}

type AnyFieldAndError struct {
	Value any
	Err   string
}

type StringFieldAndRuleset struct {
	Value   *string
	RuleSet RegExRuleset
}

// for handling calls to verifications here like VerifyBasicDBUpdate
//   - returns true if there is an error, SHOULD RETURN
//
// if there is an actual programmatic error:
//   - the real error is logged to the system and typed as apierrorkeys.FormFieldValidationError,
//   - ERROR_PROCESSING_REQUEST is retuned to the client , client should digest ERROR_PROCESSING_REQUEST
//
// if there is a validation error, the validation error message is returned to the client:
//   - the form validation message is logged to the system and typed as apierrorkeys.FormFieldUserError,
//   - the form validation message is returned to the client for disgestion
func HandleVerificationReturn(msg string, err error, r *http.Request, w http.ResponseWriter) (shouldReturn bool) {
	if err != nil {
		apierrors.HandleError(r, err, apierrorkeys.FormFieldValidationError, &apierrors.ReturnError{Msg: ERROR_PROCESSING_REQUEST, W: &w})
		shouldReturn = true
		return shouldReturn
	}
	if msg != string(DATA_OK) {
		apierrors.HandleError(r, errors.New(msg), apierrorkeys.FormFieldUserError, &apierrors.ReturnError{Msg: msg, W: &w})
		shouldReturn = true
		return shouldReturn
	}
	shouldReturn = false
	return shouldReturn

}

func IsFieldInStruct(structType interface{}, fieldName string) bool {
	val := reflect.TypeOf(structType)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Ensure we are dealing with a struct
	if val.Kind() != reflect.Struct {
		return false
	}

	for i := 0; i < val.NumField(); i++ {
		if val.Field(i).Name == fieldName {
			return true
		}
	}
	return false
}

// TrimStringFields trims leading and trailing whitespace from all string fields in a struct.
func TrimStringFields(input interface{}) {
	val := reflect.ValueOf(input)

	// Ensure we have a pointer to a struct
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return
	}

	val = val.Elem()
	//typ := val.Type()

	// Iterate over the fields of the struct
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		//fieldType := typ.Field(i)

		// Check if the field is a string
		if field.Kind() == reflect.String {
			// Trim the whitespace and set the value back to the field
			trimmed := strings.TrimSpace(field.String())
			field.SetString(trimmed)
		}
	}
}
