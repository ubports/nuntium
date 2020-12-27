package main

import (
	"reflect"
	"testing"

	"github.com/ubports/nuntium/mms"
)

func TestGetMNotificationIndPayload(t *testing.T) {
	testCases := []struct {
		args              mainFlags
		differFromDefault bool
	}{
		{},
		{mainFlags{Sender: "+12345"}, false},
		{mainFlags{Sender: "+543515924906"}, false},
		{mainFlags{SenderNotification: "+12345"}, true},
		{mainFlags{SenderNotification: "+543515924906"}, false},
	}

	for _, tc := range testCases {
		pl := getMNotificationIndPayload(tc.args)
		if !tc.differFromDefault != reflect.DeepEqual(pl, mNotificationInd) {
			differ := ""
			if !tc.differFromDefault {
				differ = "not "
			}
			t.Errorf("Push payload for args %#v should %sdiffer from default payload", tc.args, differ)
			continue
		}

		dec := mms.NewDecoder(pl)
		mni := mms.NewMNotificationInd()
		if err := dec.Decode(mni); err != nil {
			t.Errorf("Error decoding payload: %v", err)
			continue
		}

		wantFrom := tc.args.SenderNotification + "/TYPE=PLMN"
		if tc.args.SenderNotification == "" {
			wantFrom = "+543515924906/TYPE=PLMN"
		}
		if mni.From != wantFrom {
			t.Errorf("Decoded MRetrieveConf.From \"%v\" should equal %v", mni.From, wantFrom)
		}
	}
}
