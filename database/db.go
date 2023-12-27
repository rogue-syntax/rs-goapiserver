package database

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/rogue-syntax/goqb-rs"
	"github.com/rogue-syntax/rs-goapiserver/global"
	"github.com/rogue-syntax/rs-goapiserver/tls"
)

// query struct for simple where conditions
// try to add group by and go from there
type SimpleWhere struct {
	Field    string
	Operator string
	Value    string
}

// var DB *sql.DB
var DB *sqlx.DB

func StartDB() error {
	//connect to db with tls / ssl if not dev env
	if !global.EnvVars.DBTLS {
		tlsConf, err := tls.CreateTLSConf()
		if err != nil {
			return err
		}
		err = mysql.RegisterTLSConfig("custom", &tlsConf)
		if err != nil {
			return err
		}
		err = connectGDBTLS()
		if err != nil {
			return err
		}
	} else {
		err := connectGDB()
		if err != nil {
			return err
		}
	}

	return nil
}

func connectGDBTLS() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?tls=custom",
		global.EnvVars.DbserverUser,
		global.EnvVars.DbserverPW,
		global.EnvVars.Dbserver,
		global.EnvVars.DbserverPort,
		global.EnvVars.DbserverDefaultDB)

	db, dbconnerr := sqlx.Open("mysql", dsn)
	if dbconnerr != nil {
		return dbconnerr
	}
	db.SetMaxIdleConns(50)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(1 * time.Minute)
	DB = db
	//defer db.Close()
	return nil
}

func connectGDB() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		global.EnvVars.DbserverUser,
		global.EnvVars.DbserverPW,
		global.EnvVars.Dbserver,
		global.EnvVars.DbserverPort,
		global.EnvVars.DbserverDefaultDB)

	db, dbconnerr := sqlx.Open("mysql", dsn)
	if dbconnerr != nil {
		return dbconnerr
	}
	db.SetMaxIdleConns(50)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(1 * time.Minute)
	DB = db
	//defer db.Close()
	return nil
}

var allowedPostFilters = map[string]string{
	"tag_name": "t.name",
	"title":    "title",
}

func BuildFilterString(query string, filters map[string][]string, allowedFilters map[string]string) (string, []interface{}, error) {
	filterString := "WHERE 1=1"
	var inputArgs []interface{}

	// filter key is the url name of the filter used as the lookup for the allowed filters list
	for filterKey, filterValList := range filters {
		if realFilterName, ok := allowedFilters[filterKey]; ok {
			if len(filterValList) == 0 {
				continue
			}

			filterString = fmt.Sprintf("%s AND %s IN (?)", filterString, realFilterName)
			inputArgs = append(inputArgs, filterValList)
		}
	}
	// template the where clause into the original query and then expand the IN clauses with sqlx
	query, args, err := sqlx.In(fmt.Sprintf(query, filterString), inputArgs...)
	if err != nil {
		return "", nil, err
	}
	// using postgres means we need to rebind the ? bindvars that sqlx.IN creates by default to $ bindvars
	// you can omit this if you are using mysql
	//query = sqlx.Rebind(sqlx.DOLLAR, query)
	return query, args, nil
}

func GetSimpleWheres(w http.ResponseWriter, r *http.Request) (*[]SimpleWhere, error) {
	var wheres []SimpleWhere
	r.ParseForm()

	_, ok := r.PostForm["where"]
	if ok {
		for _, whereStr := range r.PostForm["where"] {
			var wh SimpleWhere
			err := json.Unmarshal([]byte(whereStr), &wh)
			if err != nil {
				return &wheres, err
			}
			wheres = append(wheres, wh)
		}
	} else {
		var wh SimpleWhere
		wh.Field = "1"
		wh.Operator = "="
		wh.Value = "1"
		wheres = append(wheres, wh)
	}
	return &wheres, nil
}

func GetFilterSqlRequest(w http.ResponseWriter, r *http.Request) (*FilterSQLRequest, error) {

	var filterReq FilterSQLRequest
	err := json.NewDecoder(r.Body).Decode(&filterReq)
	if err != nil {
		return &filterReq, err
	}
	return &filterReq, err
}

type FilterSQLRequest struct {
	Wheres *[]SimpleWhere
	Offset string
	Limit  string
}

func FilteredQuery(qb *goqb.GoQB, filterReq *FilterSQLRequest, tableViewName string, model interface{}, targ interface{}, w http.ResponseWriter, r *http.Request) (*interface{}, error) {

	wheres := filterReq.Wheres
	propModel := qb.Model(tableViewName, model)
	quer := propModel.Where((*wheres)[0].Field, (*wheres)[0].Operator, (*wheres)[0].Value)
	for i := 1; i < len(*wheres); i++ {
		quer = quer.AndWhere((*wheres)[i].Field, (*wheres)[i].Operator, (*wheres)[i].Value)
	}
	quer = quer.OffsetLimit((*filterReq).Offset, (*filterReq).Limit)
	err := quer.Get(targ)
	if err != nil {
		return &targ, err
	}

	return &targ, err

}
