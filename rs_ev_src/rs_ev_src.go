package rs_ev_src

import (
	"database/sql"
	"io"
	"os"
	"sync"
	"time"

	"encoding/gob"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// need to differentiate propagated error to application callers
//
// did the action complete successfully? ERRORFLAG_ACTION
//
// did the event stream to the file? ERRORFLAG_STREAM_EVENT_1
//
// etc
const (
	ERRORFLAG_ACTION         = "ACTION"
	ERRORFLAG_STREAM_EVENT_1 = "STREAM_1"
	ERRORFLAG_STREAM_EVENT_2 = "STREAM_2"
	ERRORFLAG_STORE_TO_SQL   = "STORE_TO_SQL"
)

// SetDBCONN sets the global variable DBCONN to the database connection
var DBCONN *sql.DB

// intialize the package with the database connection and the event names
// see the EVEventNames map in this package for an example
func INIT(db *sql.DB, eVEventNames map[EVTypes_int]string) {
	EVEventNames = eVEventNames
	DBCONN = db
}

const (
	TIMESTAMP_6   = "2006-01-02 15:04:05.999999"
	DATEFORMAT    = "2006-01-02"
	NO_REQUEST_ID = ""
)

func MAKE_TIMESTAMP_6_STR() string {
	return time.Now().Format(TIMESTAMP_6)
}

type EVEventFileLocation string

var EV_EVENT_GOB_FILE_LOCATION EVEventFileLocation = "ev_events_file.gob"
var EV_EVENT_JSONB_FILE_LOCATION EVEventFileLocation = "ev_events_file.jsonb"

type EVEventSchemaVersion float64

var EV_SCHEMA_VERSION EVEventSchemaVersion = 1.0

// CONST MAP KEYS FOR MAPPED EVENTS
type EVTypes_int int

const (
	UPDATE_TEST_DATA EVTypes_int = iota
	SOME_ACTION      EVTypes_int = iota
)

type EVScheduleTypes_int int

const (
	IMMEDIATE_ACTION EVScheduleTypes_int = iota
)

var EVEventNames = map[EVTypes_int]string{}

/*
EXAMPLE FUNCTION AND DATA WE WANT TO RUN AS AN EVENT
Some Data Types, Context Types, and Functions
we dont want to just run our function 'UpdateATestDataType' willy nilly,
we want to dispatch is as an event, with its data, and its context, and its success or failure, and log it, or queue it, or replay it , etc
We want to package these into an actionable event!
*/
type ATestDataType struct {
	id   string
	data string
}

func UpdateATestDataType(f ATestDataType) (string, error) {
	return "", nil
}

// NoMetaData:
//
// keep some empty struct around for the places where you dont need events to have meta data
//
// we've included 'NoMetaData' in this package
type NoMetaData struct {
}

/*
	THE EVENT CLASS
	Here is our Event Class and its associated types and interfaces
*/

// The action interface
// The action interface is a generic interface that can be implemented by any struct
// The struct must have a method called "Do" that takes one argument of any type and returns an error
type ISerializableEvent interface {
	ToJsonBytes() ([]byte, error)
	CheckSuccess() bool
	GetReplayType() error
	GetCalledAt() UnixTimeMilliseconds
	GetUUID() uuid.UUID
}

type IEVAction[T any, R any] interface {
	Do(data T) (R, error)
}
type UnixTimeMilliseconds int64

// The action event class
//
// EVEvent Class associates types and interfaces are:
//   - UnixTimeMilliseconds
//   - EVTypes_int
//   - IEVAction (interface)
type EVEvent[DATATYPE any, CONTEXTTYPE any, RETURNTYPE any] struct {
	Ev_id         uuid.UUID
	Ev_type       EVTypes_int
	Action_name   string
	Action        IEVAction[DATATYPE, RETURNTYPE] `gob:"-"`
	Data          DATATYPE
	MetaData      CONTEXTTYPE
	CalledAt      UnixTimeMilliseconds
	Timestamp     time.Time
	Date_time     string
	Success       bool
	Version       EVEventSchemaVersion
	ErrMsg        string
	Req_id        string //36 char uuid
	Attempt       int
	Schedule_type EVScheduleTypes_int
}

type EVEventSerial struct {
	Ev_id         uuid.UUID
	Ev_type       EVTypes_int
	Action_name   string
	Data          []uint8
	MetaData      []uint8
	CalledAt      UnixTimeMilliseconds
	Timestamp     time.Time
	Date_time     string
	Success       bool
	Version       EVEventSchemaVersion
	ErrMsg        string
	Req_id        string //36 char uuid
	Attempt       int
	Schedule_type EVScheduleTypes_int
}

// SetEVEvent must be called for every event to set the event's data, action, and metadata
//
// **IMPORTANT**:  Dont forget to add it to the register in the 'Init' function in the ev_register package if you want to be able to call this event with recalled data
//
// data and metadata must be JSON serializable using the 'encoding/json' package
//
// action must implement the IEVAction interface
//
//   - action: the action to be executed
//   - data: the data to be passed to the action
//   - metaData: the metadata to be logged along with the action
//   - mapKey: the key to map the event to
//   - scheduleType: the type of schedule for the event
//   - req_id: the request id for the event
func (e *EVEvent[DATATYPE, CONTEXTTYPE, RETURNTYPE]) SetEVEvent(action IEVAction[DATATYPE, RETURNTYPE], data *DATATYPE, metaData *CONTEXTTYPE, mapKey EVTypes_int, scheduleType EVScheduleTypes_int, req_id string) {
	e.Version = EV_SCHEMA_VERSION
	e.Action = action
	e.Data = *data
	e.MetaData = *metaData
	e.Ev_id = uuid.New()
	e.Ev_type = mapKey
	e.Action_name = EVEventNames[mapKey]
	e.CalledAt = UnixTimeMilliseconds(time.Now().UnixMilli())
	e.Req_id = req_id
	e.Schedule_type = scheduleType

	//TBD IF ADDING TO A QUEUE OR STACK?
	/*
		EVActionMap[mapKey] = func() EVEventExecError {
			err := DoEVEventAction[DATATYPE](e)
			return err
		}
	*/
	e.ErrMsg = ""
}

/*
implement the ISerializableEvent interface
*/
func (e *EVEvent[DATATYPE, CONTEXTTYPE, RETURNTYPE]) ToJsonBytes() ([]byte, error) {
	return json.Marshal(e)
}
func (e *EVEvent[DATATYPE, CONTEXTTYPE, RETURNTYPE]) CheckSuccess() bool {
	return e.Success
}
func (e *EVEvent[DATATYPE, CONTEXTTYPE, RETURNTYPE]) GetReplayType() EVTypes_int {
	return e.Ev_type
}
func (e *EVEvent[DATATYPE, CONTEXTTYPE, RETURNTYPE]) GetCalledAt() UnixTimeMilliseconds {
	return e.CalledAt
}
func (e *EVEvent[DATATYPE, CONTEXTTYPE, RETURNTYPE]) GetUUID() uuid.UUID {
	return e.Ev_id
}

/*
	func (e *EVEvent[DATATYPE, CONTEXTTYPE]) StreamEvent(ev ISerializableEvent) error {
		streamToGOBFile(ev)
		return err
	}
*/
var EVEventSlice []ISerializableEvent

type EVEventExecError struct {
	ActionError error
	StreamError error
	StoreError  error
}

// External method to call the event action
//
// **IMPORTANT**:  Dont forget to add it to the register in the 'Init' function in the ev_register package if you want to be able to call this event with recalled data
//
// # External because we cant do generic class methods
//
// This function takes an EVEvent and calls the 'Do' method of the action
//
// returns EVEventExecError:
//   - ActionError: error from the action : check for nil to see if the action was successful( nil means success )
//   - StreamError: error from streaming the event : check for nil to see if the event was streamed successfully
//   - StoreError: error from storing the event to the database : check for nil to see if the event was stored successfully
func DoEVEventAction[DATATYPE any, CONTEXTTYPE any, RETURNTYPE any](e *EVEvent[DATATYPE, CONTEXTTYPE, RETURNTYPE]) (res RETURNTYPE, evError EVEventExecError) {
	var err error = nil
	e.Timestamp = time.Now()
	e.Date_time = e.Timestamp.Format(DATEFORMAT)
	//take action
	res, err = e.Action.Do(e.Data)
	//action is not successful?
	if err != nil {
		e.Success = false
		e.ErrMsg = err.Error()
		evError.ActionError = errors.Wrap(err, ERRORFLAG_ACTION)
	} else {
		//action was succesful!
		e.Success = true
	}
	//stream the event
	eInterface := &EVEvent[interface{}, interface{}, interface{}]{
		Ev_id:       e.Ev_id,
		Ev_type:     e.Ev_type,
		Action_name: e.Action_name,
		Data:        e.Data,
		MetaData:    e.MetaData,
		CalledAt:    e.CalledAt,
		Timestamp:   e.Timestamp,
		Date_time:   e.Date_time,
		Success:     e.Success,
		Version:     e.Version,
		ErrMsg:      e.ErrMsg,
		Req_id:      e.Req_id,
		Attempt:     e.Attempt,
	}
	err = StreamEV(eInterface)
	if err != nil {
		err2 := errors.Wrap(err, ERRORFLAG_STREAM_EVENT_1)
		evError.StreamError = err2
	}
	err = StoreEV(eInterface)
	if err != nil {
		err2 := errors.Wrap(err, ERRORFLAG_STORE_TO_SQL)
		evError.StoreError = err2
	}
	return res, evError
}

// MAP OF ACTION NAMES TO ACTION HANDLERS
// HANDLERS ARE : ANONYMOUS FUNCTIONS THAT WRAP DoEVEventAction
// DoEVEventAction TAKES THE ARGUMENT TYPE AND THE EVEvent WHICH CONTAIN THE ACTION AND ITS DATA
// var EVActionMap = map[string]EVHandler{}
type EVHandler func() EVEventExecError
type EVReplayHandler func(*EVEventSerial) EVEventExecError

var EVActionMap = map[EVTypes_int]EVHandler{}

var EVReplayMap = map[EVTypes_int]EVReplayHandler{}

/*
	EXAMPLE USAGE
*/

// Here is an example of how to make an IEVAction for the event class
//
// # Make a struct that implements the IEVAction interface
//
// Make a method called 'Do' that takes the data type of the event and returns an error
//
// This is where you put the function you want to run as an event!
type ST_UPDATE_FUND struct{}

func (a ST_UPDATE_FUND) Do(atdt ATestDataType) (string, error) {
	res, err := UpdateATestDataType(atdt)
	return res, err
}

func AnExample() {

	//Somewhere in our app we have a function we need to run and some data it needs to run:
	var td ATestDataType
	//going to run this 'td' through 'UpdateATestDataType' *as an event*

	//for our event we may also want to pass some contextual metadata
	var metadata NoMetaData

	//instead of calling the function directly, we want to package it into an event
	//declare the event and its types
	var update_aTestDataType_event EVEvent[ATestDataType, NoMetaData, string]

	//set it with the action we want it to execute, the data it needs, any metadata, and the key to map it to
	//
	update_aTestDataType_event.SetEVEvent(ST_UPDATE_FUND{}, &td, &metadata, UPDATE_TEST_DATA, IMMEDIATE_ACTION, NO_REQUEST_ID)

	//Call the event to execute directly with DoEVEventAction
	DoEVEventAction[ATestDataType, NoMetaData](&update_aTestDataType_event)

	//Call the event to execute from the EVActionMap
	EVActionMap[UPDATE_TEST_DATA]()

}

var fileMutex sync.Mutex

func StreamToGOBFile(ev *EVEvent[interface{}, interface{}, interface{}]) error {
	fileMutex.Lock()
	defer fileMutex.Unlock()
	file, err := os.OpenFile(string(EV_EVENT_GOB_FILE_LOCATION), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(EVEventSlice)
	if err != nil {
		return err
	}
	return nil
}

func LoadFromGOBFile() ([]EVEvent[interface{}, interface{}, interface{}], error) {
	fileMutex.Lock()
	defer fileMutex.Unlock()
	file, err := os.Open(string(EV_EVENT_GOB_FILE_LOCATION))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var events []EVEvent[interface{}, interface{}, interface{}]
	decoder := gob.NewDecoder(file)
	for {
		var event EVEvent[interface{}, interface{}, interface{}]
		err = decoder.Decode(&event)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

func StreamEV(ev *EVEvent[interface{}, interface{}, interface{}]) error {
	fileMutex.Lock()
	defer fileMutex.Unlock()
	file, err := os.OpenFile(string(EV_EVENT_JSONB_FILE_LOCATION), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(ev)
	if err != nil {
		return err
	}
	return nil
}

func LoadFromJSONFile() ([]EVEventSerial, error) {
	fileMutex.Lock()
	defer fileMutex.Unlock()
	file, err := os.Open(string(EV_EVENT_JSONB_FILE_LOCATION))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var events []EVEventSerial
	decoder := json.NewDecoder(file)
	for {
		var ev EVEventSerial
		if err := decoder.Decode(&ev); err == io.EOF {
			break
		} else if err != nil {
			// handle error
		}
		events = append(events, ev)
	}

	return events, nil
}

func StoreEV(e *EVEvent[interface{}, interface{}, interface{}]) error {
	dataJson, err := json.Marshal(&e.Data)
	if err != nil {
		return err
	}

	metaDataJson, err := json.Marshal(&e.MetaData)
	if err != nil {
		return err
	}

	_, err = DBCONN.Exec(`
			INSERT INTO EVEvents (Ev_id, Ev_type, Action_name, Data, MetaData, CalledAt, Timestamp, Date_time, Success, Version, ErrMsg, Req_id, Attempt, Schedule_type)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,?)`,
		e.Ev_id,
		e.Ev_type,
		e.Action_name,
		dataJson,
		metaDataJson,
		e.CalledAt,
		e.Timestamp, //this one is for timestamp
		e.Date_time, //this one is for datetime
		e.Success,
		e.Version,
		e.ErrMsg,
		e.Req_id,
		e.Attempt,
		e.Schedule_type,
	)

	return err
}

func streamEVEvent(ev EVEvent[interface{}, interface{}, interface{}]) error {

	return nil

}

var SQL string = `
CREATE TABLE EVEvents (
    Ev_id char(36),
	Ev_type INT,
	Action_name CHAR(100),
	Data JSON,
	MetaData JSON,
    CalledAt BIGINT,
	Timestamp TIMESTAMP(6),
    Date_time DATETIME(6),
	Success BOOLEAN,
	Version DECIMAL(5,3),
	ErrMsg CHAR(255),
	Req_id char(36),
	Attempt TINYINT DEFAULT 0,
	Schedule_type tinyint(4) NOT NULL DEFAULT 0,
  PRIMARY KEY (Ev_id, Date_time)
) ENGINE = INNODB,
  CHARACTER SET utf8mb4,
  COLLATE utf8mb4_general_ci
PARTITION BY RANGE (MONTH(Date_time)) (
	PARTITION p0 VALUES LESS THAN (1),
	PARTITION p1 VALUES LESS THAN (4),
	PARTITION p2 VALUES LESS THAN (8),
	PARTITION p3 VALUES LESS THAN (12),
	PARTITION p4 VALUES LESS THAN MAXVALUE
);
`
