package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/pkg/errors"

	"github.com/rogue-syntax/rs-goapiserver/apireturn/apierrorkeys"
)

func Marshal(i interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(i)
	return bytes.TrimRight(buffer.Bytes(), "\n"), err
}

type Recaptcha3Req struct {
	Secret   string `json:"secret"`
	Response string `json:"response"`
}

type ReqHeader struct {
	HeaderName  string
	HeaderValue string
}

func HttpPostReq(method string, payload interface{}, url string, reqHeaders []ReqHeader, addHeaders []ReqHeader) (error, []byte) {
	if reqHeaders == nil {
		defaultHeader := []ReqHeader{
			{HeaderName: "Content-Type", HeaderValue: "application/json; charset=utf-8"},
			{HeaderName: "Accept", HeaderValue: "application/json"},
		}
		reqHeaders = defaultHeader
	}
	if addHeaders != nil {
		reqHeaders = append(reqHeaders, addHeaders...)
	}
	var returnByes []byte
	var reqBytes []byte
	var err error
	if payload != nil {
		reqBytes, err = json.Marshal(&payload)
		if err != nil {
			return err, returnByes
		}
	}

	request, err := http.NewRequest(method, url, bytes.NewBuffer(reqBytes))

	for i := 0; i < len(reqHeaders); i++ {
		request.Header.Set(reqHeaders[i].HeaderName, reqHeaders[i].HeaderValue)
	}
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return err, returnByes
	}
	defer response.Body.Close()
	rBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err, returnByes
	}

	return nil, rBody
}

func GetReqFromJSON(r *http.Request, reqObj interface{}) error {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&reqObj)
	if err != nil {
		return err
	}
	return nil
}

func CheckForPointer(i interface{}) bool {
	t := reflect.TypeOf(i)
	if t.Kind() == reflect.Ptr {
		return true
	}
	return false
}

func DistilInterfaceString(v interface{}) (returnStr string, err error) {
	if v == nil {
		returnStr = ""
		return returnStr, err
	}

	switch v.(type) {
	case *string:
		returnStr = *v.(*string)
	case *int:
		returnStr = fmt.Sprint(*v.(*int))
		// add cases for other types if needed
	default:
		err = errors.New(apierrorkeys.InvalidType)
	}

	return returnStr, err
}

/*



	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		apierrors.HandleApiReqErrors(err, apiErrorKeys.ys.ys.APIReqError, usr, w, r)
		return
	}
	defer response.Body.Close()
	rBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		apierrors.HandleApiReqErrors(err, apiErrorKeys.APIReqError, usr, w, r)
		return
	}
*/
