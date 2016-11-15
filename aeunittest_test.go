package aeunittest

import (
	"testing"
)

func Test1(t *testing.T) {
	tcs := TestCases{}

	err := tcs.Load(`test\lifelog test cases - Goal.csv`, ',', true)
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range tcs {
		t.Log(i, tc)
	}

}
