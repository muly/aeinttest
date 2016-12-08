package aeunittest

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	//"fmt"

	gorillacontext "github.com/gorilla/context"
	"golang.org/x/net/context"
)

type (
	TestCase struct {
		Name             string
		RequestBody      string
		HttpVerb         string
		Uri              string
		WantStatusCode   int
		WantResponseBody string
		Skip             bool
		//
		*testing.T
		context.Context
		http.Handler
	}

	TestCases []TestCase
)

func (tc TestCase) Run1() (ResponseBody []byte) { // compare response code only and return response body received

	if tc.Skip {
		tc.Log("Skipped test case:", tc.Name)
		return
	}

	// prepare request
	body := strings.NewReader(tc.RequestBody)
	req, err := http.NewRequest(tc.HttpVerb, tc.Uri, body) //inst.NewRequest("GET", goalUrl, body) //
	if err != nil {
		tc.Error(err)
	}

	gorillacontext.Set(req, "Context", tc.Context)

	// prepare response writer
	record := httptest.NewRecorder()

	// make the request
	tc.ServeHTTP(record, req)

	//tc.Log(record.Body)

	if got := record.Code; tc.WantStatusCode != got {
		tc.Error(tc.Name, ": Status Code: wanted ", tc.WantStatusCode, " but got ", got)
	}

	return record.Body.Bytes()

}

func (tc TestCase) Run() {

	if tc.Skip {
		tc.Log("Skipped test case:", tc.Name)
		return
	}

	g1 := tc.Run1()
	var got interface{}

	if err := json.Unmarshal(g1, &got); err != nil {
		tc.Error("Want Response Body invalid format: ", err.Error())
	}

	var want interface{}

	if err := json.Unmarshal([]byte(tc.WantResponseBody), &want); err != nil {
		tc.Error("Want Response Body invalid format: ", err.Error())
	}

	//tc.Log("Want:", want)
	//tc.Log("Got: ", got)
	//tc.Log("DeepEqual:", reflect.DeepEqual(got, want))

	if !reflect.DeepEqual(got, want) {
		tc.Error(tc.Name, ": Response Body : wanted ", tc.WantResponseBody, " but got ", string(g1))
	}

}

// Load method loads the given test cases data from the flat file into the TestCases slice object.
// if header is provided, will parse the data accordign the header, otherwise, the default order will be used.
func (tcs *TestCases) Load(filePath string, delim rune, hasHeader bool) error {

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	r := csv.NewReader(file)
	r.Comma = delim
	r.Comment = '#'

	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	var name, requestBody, httpVerb, uri, wantStatusCode, wantResponseBody, skip int

	for i, row := range records {

		if i == 0 && hasHeader == true {
			for j, c := range row {
				//find the position of each field
				//fmt.Println("###", j, c)
				switch c {
				case "Name":
					name = j
				case "RequestBody":
					requestBody = j
				case "HttpVerb":
					httpVerb = j
				case "Uri":
					uri = j
				case "WantStatusCode":
					wantStatusCode = j
				case "WantResponseBody":
					wantResponseBody = j
				case "Skip":
					skip = j
				}
			}
			continue

		} else if i == 0 {
			name = 0
			requestBody = 1
			httpVerb = 2
			uri = 3
			wantStatusCode = 4
			wantResponseBody = 5
			skip = 6
		}

		if len(row[name])+len(row[requestBody])+len(row[httpVerb])+len(row[uri])+len(row[wantStatusCode])+
			len(row[wantResponseBody])+len(row[skip]) == 0 {
			continue // if all the fields are blank, then skip
		} else if len(row[name]) == 0 || len(row[httpVerb]) == 0 || len(row[uri]) == 0 || len(row[wantStatusCode]) == 0 {
			return errors.New("Missing manditory information in row " + strconv.Itoa(i) + " (" + row[name] + ")") // if only manditory fields are blank, then error out
		}

		tc := TestCase{}
		tc.Name = row[name]
		tc.RequestBody = row[requestBody]
		tc.HttpVerb = row[httpVerb]
		tc.Uri = row[uri]
		tc.WantResponseBody = row[wantResponseBody]
		switch strings.ToUpper(row[skip]) {
		case "YES", "TRUE", "1":
			tc.Skip = true
		case "NO", "FALSE", "0":
			tc.Skip = false
		}

		statusCode, err := strconv.Atoi(row[wantStatusCode][0:3]) // take the first 3 digits of the status code. skip the rest, the information that user might put like in '200 OK' or '400 Bad Request'
		if err != nil {
			return err
		}
		tc.WantStatusCode = statusCode
		*tcs = append(*tcs, tc)
	}

	return nil
}
