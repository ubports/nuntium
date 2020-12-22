package main

import (
	"reflect"
	"testing"
)

func TestGetMRetrieveConfPayload(t *testing.T) {
	testCases := []struct {
		args              mainFlags
		differFromDefault bool
	}{
		{},
		{mainFlags{Sender: "+12345"}, true},
		{mainFlags{Sender: "01189998819991197253"}, false},
	}

	for _, tc := range testCases {
		pl := GetMRetrieveConfPayload(tc.args)
		if !tc.differFromDefault != reflect.DeepEqual(pl, mRetrieveConf) {
			differ := ""
			if !tc.differFromDefault {
				differ = "not "
			}
			t.Errorf("Payload for args %v should %sdiffer from default payload", tc.args, differ)
		}
	}
}
