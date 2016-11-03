package aeunittest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
