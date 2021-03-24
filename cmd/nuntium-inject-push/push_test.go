package main

import (
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

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
		{mainFlags{TransactionId: "12345abcde"}, true},
		{mainFlags{TransactionId: ""}, false},
		{mainFlags{ErrorActivateContext: 1}, true},
		{mainFlags{ErrorGetProxy: 5}, true},
		{mainFlags{ErrorDownloadStorage: 9}, true},
		{mainFlags{ErrorReceiveHandle: 1}, true},
		{mainFlags{ErrorReceiveStorage: 1}, true},
		{mainFlags{ErrorRespondHandle: 1}, true},
		{mainFlags{ErrorRespondStorage: 1}, true},
		{mainFlags{ErrorActivateContext: 1, ErrorGetProxy: 1, ErrorDownloadStorage: 1, ErrorReceiveHandle: 1, ErrorReceiveStorage: 1, ErrorRespondHandle: 1, ErrorRespondStorage: 1}, true},
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
		mni := mms.NewMNotificationInd(time.Time{})
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

		if mni.TransactionId != tc.args.TransactionId {
			t.Errorf("Decoded MRetrieveConf.TransactionId \"%v\" should equal %v", mni.TransactionId, tc.args.TransactionId)
		}

		if cl, err := url.Parse(mni.ContentLocation); err != nil {
			t.Errorf("Couldn't parse MRetrieveConf.ContentLocation \"%s\": %v", mni.ContentLocation, err)
		} else {
			wantCLPrefix := "http://localhost:9191/mms"
			if !strings.HasPrefix(cl.String(), wantCLPrefix) {
				t.Errorf("Decoded MNotificationInd.ContentLocation \"%s\" should start with \"%s\"", mni.ContentLocation, wantCLPrefix)
			}
			if tc.args.ErrorActivateContext > 0 {
				if ui64, err := strconv.ParseUint(cl.Query().Get(mms.DebugErrorActivateContext), 10, 64); err != nil {
					t.Errorf("Couldn't parse \"%s\": %s", mms.DebugErrorActivateContext, err)
				} else if tc.args.ErrorActivateContext != ui64 {
					t.Errorf("Decoded MNotificationInd.ContentLocation query parameter \"%s\" is %d, want %d", mms.DebugErrorActivateContext, ui64, tc.args.ErrorActivateContext)
				}
			}
			if tc.args.ErrorGetProxy > 0 {
				if ui64, err := strconv.ParseUint(cl.Query().Get(mms.DebugErrorGetProxy), 10, 64); err != nil {
					t.Errorf("Couldn't parse \"%s\": %s", mms.DebugErrorGetProxy, err)
				} else if tc.args.ErrorGetProxy != ui64 {
					t.Errorf("Decoded MNotificationInd.ContentLocation query parameter \"%s\" is %d, want %d", mms.DebugErrorGetProxy, ui64, tc.args.ErrorGetProxy)
				}
			}
			if tc.args.ErrorDownloadStorage > 0 {
				if ui64, err := strconv.ParseUint(cl.Query().Get(mms.DebugErrorDownloadStorage), 10, 64); err != nil {
					t.Errorf("Couldn't parse \"%s\": %s", mms.DebugErrorDownloadStorage, err)
				} else if tc.args.ErrorDownloadStorage != ui64 {
					t.Errorf("Decoded MNotificationInd.ContentLocation query parameter \"%s\" is %d, want %d", mms.DebugErrorDownloadStorage, ui64, tc.args.ErrorDownloadStorage)
				}
			}
			if tc.args.ErrorReceiveHandle > 0 {
				if ui64, err := strconv.ParseUint(cl.Query().Get(mms.DebugErrorReceiveHandle), 10, 64); err != nil {
					t.Errorf("Couldn't parse \"%s\": %s", mms.DebugErrorReceiveHandle, err)
				} else if tc.args.ErrorReceiveHandle != ui64 {
					t.Errorf("Decoded MNotificationInd.ContentLocation query parameter \"%s\" is %d, want %d", mms.DebugErrorReceiveHandle, ui64, tc.args.ErrorReceiveHandle)
				}
			}
			if tc.args.ErrorReceiveStorage > 0 {
				if ui64, err := strconv.ParseUint(cl.Query().Get(mms.DebugErrorReceiveStorage), 10, 64); err != nil {
					t.Errorf("Couldn't parse \"%s\": %s", mms.DebugErrorReceiveStorage, err)
				} else if tc.args.ErrorReceiveStorage != ui64 {
					t.Errorf("Decoded MNotificationInd.ContentLocation query parameter \"%s\" is %d, want %d", mms.DebugErrorReceiveStorage, ui64, tc.args.ErrorReceiveStorage)
				}
			}
			if tc.args.ErrorRespondHandle > 0 {
				if ui64, err := strconv.ParseUint(cl.Query().Get(mms.DebugErrorRespondHandle), 10, 64); err != nil {
					t.Errorf("Couldn't parse \"%s\": %s", mms.DebugErrorRespondHandle, err)
				} else if tc.args.ErrorRespondHandle != ui64 {
					t.Errorf("Decoded MNotificationInd.ContentLocation query parameter \"%s\" is %d, want %d", mms.DebugErrorRespondHandle, ui64, tc.args.ErrorRespondHandle)
				}
			}
			if tc.args.ErrorRespondStorage > 0 {
				if ui64, err := strconv.ParseUint(cl.Query().Get(mms.DebugErrorRespondStorage), 10, 64); err != nil {
					t.Errorf("Couldn't parse \"%s\": %s", mms.DebugErrorRespondStorage, err)
				} else if tc.args.ErrorRespondStorage != ui64 {
					t.Errorf("Decoded MNotificationInd.ContentLocation query parameter \"%s\" is %d, want %d", mms.DebugErrorRespondStorage, ui64, tc.args.ErrorRespondStorage)
				}
			}
		}
	}
}
