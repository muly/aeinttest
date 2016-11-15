package aeunittest

import (
	"encoding/csv"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	//"fmt"

	gorillacontext "github.com/gorilla/context"
	"golang.org/x/net/context"
)

type (
	TestCase struct {
		Name           string
		RequestBody    string
		HttpVerb       string
		Uri            string
		WantStatusCode int
		//
		*testing.T
		context.Context
		http.Handler
	}

	TestCases []TestCase
)

func (tc TestCase) Run() {
	// prepare request
	body := strings.NewReader(tc.RequestBody)
	req, err := http.NewRequest(tc.HttpVerb, tc.Uri, body) //inst.NewRequest("GET", goalUrl, body) //
	if err != nil {
		tc.T.Error(err)
	}
	gorillacontext.Set(req, "Context", tc.Context)

	// prepare response writer
	record := httptest.NewRecorder()

	// make the request
	tc.Handler.ServeHTTP(record, req)

	got := record.Code

	if tc.WantStatusCode != got {
		tc.T.Error(tc.Name, ": Status Code: wanted ", tc.WantStatusCode, " but got ", got)
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
	//r.FieldsPerRecord = 5 // 5 indicates the 5 test case related fields of the TestCase struct

	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	var name, requestBody, httpVerb, uri, wantStatusCode int

	//fmt.Println("########################## 1")

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
				}
			}
			continue

		} else if i == 0 {
			name = 0
			requestBody = 1
			httpVerb = 2
			uri = 3
			wantStatusCode = 4
		}

		if len(row[name])+len(row[requestBody])+len(row[httpVerb])+len(row[uri])+len(row[wantStatusCode]) == 0 {
			continue
		} else if len(row[name]) == 0 || len(row[httpVerb]) == 0 || len(row[uri]) == 0 || len(row[wantStatusCode]) == 0 {
			return errors.New("Missing manditory information in row " + strconv.Itoa(i))
		}

		tc := TestCase{}
		tc.Name = row[name]
		tc.RequestBody = row[requestBody]
		tc.HttpVerb = row[httpVerb]
		tc.Uri = row[uri]

		//fmt.Println("row:", row)
		//fmt.Println(name, requestBody, httpVerb, uri, wantStatusCode)
		//fmt.Println("wantStatusCode:", row[wantStatusCode][0:3])

		statusCode, err := strconv.Atoi(row[wantStatusCode][0:3])
		if err != nil {
			return err
		}
		tc.WantStatusCode = statusCode
		*tcs = append(*tcs, tc)
	}

	return nil
}
