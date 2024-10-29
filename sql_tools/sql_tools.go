package sql_tools

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/rogue-syntax/rs-goapiserver/apierrors"
	"github.com/rogue-syntax/rs-goapiserver/apireturn/apierrorkeys"
	"github.com/rogue-syntax/rs-goapiserver/database"
)

const (
	ORDERBY_ASC  = "ASC"
	ORDERBY_DESC = "DESC"
)

var ORDERBY_CODE_INPUT = map[int]string{-1: ORDERBY_DESC, 1: ORDERBY_ASC}
var ORDERBY_CODE_OUTPUT = map[string]int{ORDERBY_ASC: 1, ORDERBY_DESC: -1}

var GetOrderByCode = func(orderInt int) string {
	if val, ok := ORDERBY_CODE_INPUT[orderInt]; ok {
		return val
	}
	return ORDERBY_ASC
}

var VerifySortColFromStringAndStruct = func(sortCol string, data interface{}) bool {
	val := reflect.TypeOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem() // If a pointer to a struct is passed, get the type of the dereferenced object
	}

	// Check if the struct has the field
	for i := 0; i < val.NumField(); i++ {
		if val.Field(i).Name == sortCol {
			return true
		}
	}
	return false
}

/*
func CreateUpdateStatement(data interface{}, tableName string, excludeFields []string) (string, error) {

		// Get the type of the struct
		dataType := reflect.TypeOf(data)

		// Create the UPDATE statement
		updateStmt := fmt.Sprintf("UPDATE %s SET ", tableName)

		// Iterate over the fields of the struct
		for i := 0; i < dataType.NumField(); i++ {
			field := dataType.Field(i)

			// Check if the field should be excluded
			if Contains(excludeFields, field.Name) {
				continue
			}

			// Add the field to the UPDATE statement
			updateStmt += fmt.Sprintf("%s = ?, ", field.Name)
		}

		// Remove the trailing comma and space
		updateStmt = strings.TrimSuffix(updateStmt, ", ")

		// Add the WHERE clause
		updateStmt += "WHERE id = ?"

		return updateStmt, nil
	}
*/
func Contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

func CreateUpdateStatement(data interface{}, tableName string, excludeFields []string) (string, []interface{}, error) {
	// Get the type of the struct
	dataType := reflect.TypeOf(data)

	// Create the UPDATE statement
	updateStmt := fmt.Sprintf("UPDATE %s SET ", tableName)

	// Slice to store the values of the fields
	values := []interface{}{}

	// Iterate over the fields of the struct
	for i := 0; i < dataType.NumField(); i++ {
		field := dataType.Field(i)

		// Check if the field should be excluded
		if Contains(excludeFields, field.Name) {
			continue
		}

		// Add the field to the UPDATE statement
		updateStmt += fmt.Sprintf("%s = ?, ", field.Name)

		// Get the value of the field
		fieldValue := reflect.ValueOf(data).Field(i).Interface()

		// Append the value to the slice
		values = append(values, fieldValue)
	}

	// Remove the trailing comma and space
	updateStmt = strings.TrimSuffix(updateStmt, ", ")

	// Add the WHERE clause
	updateStmt += " "
	//updateStmt += "WHERE id = ?"

	return updateStmt, values, nil
}

func CreateInsertStatement(data interface{}, tableName string, excludeFields []string, includeMap map[string]string) (string, []interface{}, error) {
	// Get the type of the struct
	dataType := reflect.TypeOf(data)

	// Create the INSERT statement
	insertStmt := fmt.Sprintf("INSERT INTO %s (", tableName)

	// Slice to store the values of the fields
	values := []interface{}{}

	// Iterate over the fields of the struct
	for i := 0; i < dataType.NumField(); i++ {
		field := dataType.Field(i)

		// Check if the field should be excluded
		if Contains(excludeFields, field.Name) {
			continue
		}

		val, ok := includeMap[field.Name]
		if ok {
			// Add the field name to the INSERT statement
			insertStmt += fmt.Sprintf("%s, ", field.Name)

			// Get the value of the field
			fieldValue := val

			// Append the value to the slice
			values = append(values, fieldValue)
		} else {
			// Add the field name to the INSERT statement
			insertStmt += fmt.Sprintf("%s, ", field.Name)

			// Get the value of the field
			fieldValue := reflect.ValueOf(data).Field(i).Interface()

			// Append the value to the slice
			values = append(values, fieldValue)
		}

	}

	// Remove the trailing comma and space
	insertStmt = strings.TrimSuffix(insertStmt, ", ")

	// Add the VALUES clause
	insertStmt += ") VALUES ("
	for range values {
		insertStmt += "?, "
	}
	insertStmt = strings.TrimSuffix(insertStmt, ", ")
	insertStmt += ")"

	return insertStmt, values, nil
}

type Comparitor string

const (
	Equal            Comparitor = "="
	NotEqual         Comparitor = "!="
	GreaterThan      Comparitor = ">"
	LessThan         Comparitor = "<"
	GreaterThanEqual Comparitor = ">="
	LessThanEqual    Comparitor = "<="
	Like             Comparitor = "LIKE"
	NotLike          Comparitor = "NOT LIKE"
)

type AndOr string

const (
	And AndOr = "AND"
	Or  AndOr = "OR"
)

type SimpleQueryComparison struct {
	AndOr      AndOr
	Field      string
	Value      string
	Comparator Comparitor
}

// return bool if string is taken in a table
//   - table: table to check
//   - value: value to check
//   - field: field to check
//   - exclusionField: field to exclude i.e. Fund_id
//   - exclusionValue: value to exclude i.e. 1
//
// exclusion should be inclusion
// {feild, value, comparator }
func IsStringTaken(table string, value string, field string, constraints []SimpleQueryComparison) (bool, error) {
	var count int
	var valuesSli []interface{}

	qStr := "SELECT COUNT(*) FROM " + table + " WHERE " + field + " = ?"
	valuesSli = append(valuesSli, value)

	if constraints != nil && len(constraints) > 0 {
		for i := 0; i < len(constraints); i++ {
			//EMPTY STRING CHECK
			if strings.TrimSpace(constraints[i].Value) != "" {
				qStr += " " + string(constraints[i].AndOr) + " " + constraints[i].Field + " " + string(constraints[i].Comparator) + " ? "
				valuesSli = append(valuesSli, constraints[i].Value)
			}

		}
	}
	qStr += ";"
	err := database.DB.Get(&count, qStr, valuesSli...)
	logme := apierrors.NewLogError(apierrorkeys.DBQueryError, apierrors.LogJsonArray(qStr, table, value, valuesSli, field))
	fmt.Println(logme)
	if err != nil {
		jsonError := apierrors.LogJsonArray(qStr, table, value, field)
		return false, errors.Wrap(err, apierrors.NewLogError(apierrorkeys.DBQueryError, jsonError))
	}

	return count > 0, nil
}

func SimpleRefCount(table string, value string, field string, contraints []SimpleQueryComparison) (int, error) {
	var count int
	var valuesSli []interface{}

	qStr := "SELECT COUNT(*) FROM " + table + " WHERE " + field + " = ?"
	valuesSli = append(valuesSli, value)

	if contraints != nil && len(contraints) > 0 {
		for i := 0; i < len(contraints); i++ {
			qStr += " " + string(contraints[i].AndOr) + " " + contraints[i].Field + " " + string(contraints[i].Comparator) + " ?"
			valuesSli = append(valuesSli, contraints[i].Value)
		}
	}
	qStr += ";"
	err := database.DB.Get(&count, qStr, valuesSli...)
	if err != nil {
		jsonError := apierrors.LogJsonArray(qStr, table, value, field)
		return count, errors.Wrap(err, apierrors.NewLogError(apierrorkeys.DBQueryError, jsonError))
	}

	return count, nil
}

var MYSQL_COMMANDS = []string{
	"SELECT", "INSERT", "UPDATE", "DELETE", "REPLACE", "LOAD DATA", "LOAD XML",
	"CREATE", "ALTER", "DROP", "TRUNCATE", "RENAME", "SHOW", "DESCRIBE", "EXPLAIN",
	"LOCK", "UNLOCK", "SET", "USE", "START TRANSACTION", "COMMIT", "ROLLBACK",
	"SAVEPOINT", "RELEASE SAVEPOINT", "SET TRANSACTION", "GRANT", "REVOKE",
}

func CheckForSQLInjection(s string) bool {
	// Check for common SQL injection patterns
	sqlInjectionPatterns := []string{
		`(?i)\bSELECT\b`, `(?i)\bINSERT\b`, `(?i)\bUPDATE\b`, `(?i)\bDELETE\b`,
		`(?i)\bDROP\b`, `(?i)\bTRUNCATE\b`, `(?i)\bALTER\b`, `(?i)\bCREATE\b`,
		`(?i)\bREPLACE\b`, `(?i)\bEXEC\b`, `(?i)\bUNION\b`, `(?i)\bOR\b`,
		`(?i)\bAND\b`, `(?i)\b--\b`, `(?i)\b;\b`, `(?i)\b'\b`, `(?i)\b"\b`,
	}

	for _, pattern := range sqlInjectionPatterns {
		matched, _ := regexp.MatchString(pattern, s)
		if matched {
			return true
		}
	}

	// Check if the string matches any of the MySQL commands exactly
	for _, command := range MYSQL_COMMANDS {
		if strings.EqualFold(s, command) {
			return true
		}
	}

	return false
}

type UniqueId struct {
	Unique_id_value      interface{}
	Unique_id_field_name string
	AndOr                AndOr
	Compartitor          Comparitor
}

type TableParams struct {
	OrderBy   string
	Direction string
	Offset    int
	Limit     int
}

func MakeTableQuery(baseQuery string, uniqueIds []UniqueId, tableParams *TableParams) (string, []interface{}, error) {
	var valuesSli []interface{}

	var SqlInjected bool
	if CheckForSQLInjection(tableParams.OrderBy) || CheckForSQLInjection(tableParams.Direction) {
		SqlInjected = true
	}

	if SqlInjected {
		return "", nil, errors.New("SQL Injection detected: " + tableParams.OrderBy + "," + tableParams.Direction)
	}

	qStr := baseQuery

	if uniqueIds != nil && len(uniqueIds) > 0 {
		qStr += " WHERE "
		for i := 0; i < len(uniqueIds); i++ {
			qStr += uniqueIds[i].Unique_id_field_name + " " + string(uniqueIds[i].Compartitor) + " ? "
			valuesSli = append(valuesSli, uniqueIds[i].Unique_id_value)
			if i < len(uniqueIds)-1 {
				qStr += " " + string(uniqueIds[i].AndOr) + " "
			}
		}
	}

	if tableParams.OrderBy != "" {
		qStr += " ORDER BY " + tableParams.OrderBy + " " + tableParams.Direction
	}

	if tableParams.Limit > 0 {
		qStr += " LIMIT ?"
		valuesSli = append(valuesSli, tableParams.Limit)
	}

	if tableParams.Offset > 0 {
		qStr += " OFFSET ?"
		valuesSli = append(valuesSli, tableParams.Offset)
	}

	return qStr, valuesSli, nil
}
