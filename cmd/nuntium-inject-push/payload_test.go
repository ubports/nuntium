package main

import (
	"reflect"
	"testing"

	"github.com/ubports/nuntium/mms"
)

func TestGetMRetrieveConfPayload(t *testing.T) {
	testCases := []struct {
		args              mainFlags
		differFromDefault bool
	}{
		{},
		{mainFlags{Sender: "+12345"}, true},
		{mainFlags{Sender: "01189998819991197253"}, false},
		{mainFlags{SenderNotification: "+12345"}, false},
		{mainFlags{SenderNotification: "01189998819991197253"}, false},
		{mainFlags{TransactionId: "12345abcde"}, true},
		{mainFlags{TransactionId: ""}, false},
		{mainFlags{TransactionId: "123456789"}, false},
	}

	for _, tc := range testCases {
		pl := getMRetrieveConfPayload(tc.args)
		if !tc.differFromDefault != reflect.DeepEqual(pl, mRetrieveConf) {
			differ := ""
			if !tc.differFromDefault {
				differ = "not "
			}
			t.Errorf("Payload for args %#v should %sdiffer from default payload", tc.args, differ)
			continue
		}

		dec := mms.NewDecoder(pl)
		mrc := mms.NewMRetrieveConf("testUUID")
		if err := dec.Decode(mrc); err != nil {
			t.Errorf("Error decoding payload: %v", err)
			continue
		}

		wantFrom := "01189998819991197253/TYPE=PLMN"
		if tc.args.Sender != "" {
			wantFrom = tc.args.Sender + "/TYPE=PLMN"
		}
		if mrc.From != wantFrom {
			t.Errorf("Decoded MRetrieveConf.From \"%v\" should equal \"%v\"", mrc.From, wantFrom)
		}

		wantTransaionId := "123456789"
		if tc.args.TransactionId != "" {
			wantTransaionId = tc.args.TransactionId
		}

		if mrc.TransactionId != wantTransaionId {
			t.Errorf("Decoded MRetrieveConf.TransactionId \"%v\" should equal \"%v\"", mrc.TransactionId, wantTransaionId)
		}
	}
}
