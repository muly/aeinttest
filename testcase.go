// Package aeinttest is the Integration test library to support writing automated test code for your
// golang API code for those that will be deployed on AppEngine

//

package aeinttest

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"golang.org/x/net/context"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	gorillacontext "github.com/gorilla/context"
)

type (
	TestCase struct {
		Name             string
		RequestBody      string
		HttpVerb         string
		Uri              string
		WantStatusCode   int
		WantResponseBody string
		SkipFlag         bool
		//
		*testing.T
		context.Context
		http.Handler
	}

	TestCases []TestCase
)

// Method RunCheckStatusCode, for the given test case record, invokes the call to the API on the given test server,
// compares want response status code to the got.
// This method doesn't check the response body, but will return the got response body to the caller.
// Caller is responsible to test for the response body.
// This limitation is helpful in situations where the response body cannot be directly tested,
// for example, if the response body contains undeterministic fields like created date, modified date.
// If that is not the case, use the RunCase() method instead.
func (tc TestCase) RunCheckStatusCode() (responseBody []byte) {
	// prepare response writer
	record := httptest.NewRecorder()

	//tc.Run(tc.Name, func(t *testing.T) { //Note: removed the code related to Subtests as it looks like Appengine SDK is not yet using Go 1.7 as of 2016 Dec.
	if tc.SkipFlag {
		tc.Log("Skipped test case:", tc.Name)
		return
	}

	// prepare request
	body := strings.NewReader(tc.RequestBody)
	req, err := http.NewRequest(tc.HttpVerb, tc.Uri, body)
	if err != nil {
		tc.Error(err)
	}

	gorillacontext.Set(req, "Context", tc.Context)

	// make the request
	tc.ServeHTTP(record, req)

	if got := record.Code; tc.WantStatusCode != got {
		tc.Error(tc.Name, ": Status Code: wanted ", tc.WantStatusCode, " but got ", got)
	}
	//})
	return record.Body.Bytes()
}

// Method RunCase does same thing as RunCheckStatusCode.
// The major difference is that, instead of returning the (got) response body,
// RunCase compares the want response body with the got, and pass/fail the test case accordingly.
//
// The order of elements in the json (between got and want) is not strict,
// however, all the fields in the want should exist in got (including the empty fields), and vice versa.
// otherwise the test fails.
func (tc TestCase) RunCase() {
	//tc.Run(tc.Name, func(t *testing.T) {
	if tc.SkipFlag {
		tc.Log("Skipped test case:", tc.Name)
		//return
	}

	// execute test case to check the status code and capture the response body
	gotResponseBody := tc.RunCheckStatusCode()

	// compare the 'got' with 'want', and report if not matching

	var got interface{}
	if err := json.Unmarshal(gotResponseBody, &got); err != nil {
		tc.Error(tc.Name, ": Got Response Body invalid format: \n", string(gotResponseBody), "\n", err.Error())
	}

	var want interface{}
	if err := json.Unmarshal([]byte(tc.WantResponseBody), &want); err != nil {
		tc.Error(tc.Name, ": Want Response Body invalid format: \n", tc.WantResponseBody, "\n", err.Error())
	}

	if !reflect.DeepEqual(got, want) {
		tc.Error(tc.Name, ": Response Body : wanted ", tc.WantResponseBody, " but got ", string(gotResponseBody))
	}
	//})
}

// Load method loads the given test cases data from the flat file into the TestCases slice object.
// If header is provided, will parse the data accordign the header, otherwise, the default order will be used.
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
			return errors.New("Missing manditory information in row " + strconv.Itoa(i) + " (" + row[name] + ")") // if any manditory fields are blank, then error out
		}

		tc := TestCase{}
		tc.Name = row[name]
		tc.RequestBody = row[requestBody]
		tc.HttpVerb = row[httpVerb]
		tc.Uri = row[uri]
		tc.WantResponseBody = row[wantResponseBody]
		switch strings.ToUpper(row[skip]) {
		case "YES", "TRUE", "1":
			tc.SkipFlag = true
		case "NO", "FALSE", "0":
			tc.SkipFlag = false
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
