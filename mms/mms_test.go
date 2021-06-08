/*
 * Copyright 2014 Canonical Ltd.
 *
 * Authors:
 * Sergio Schvezov: sergio.schvezov@cannical.com
 *
 * This file is part of mms.
 *
 * mms is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; version 3.
 *
 * mms is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package mms

import (
	"net/url"
	"reflect"
	"testing"
	"time"

	. "launchpad.net/gocheck"
)

type MMSTestSuite struct{}

var _ = Suite(&MMSTestSuite{})

func (s *MMSTestSuite) TestNewMSendReq(c *C) {
	recipients := []string{"+11111", "+22222", "+33333"}
	expectedRecipients := []string{"+11111/TYPE=PLMN", "+22222/TYPE=PLMN", "+33333/TYPE=PLMN"}
	mSendReq := NewMSendReq(recipients, []*Attachment{}, false)
	c.Check(mSendReq.To, DeepEquals, expectedRecipients)
	c.Check(mSendReq.ContentType, Equals, "application/vnd.wap.multipart.related")
	c.Check(mSendReq.Type, Equals, byte(TYPE_SEND_REQ))
}

func TestMNotificationInd_Expire(t *testing.T) {
	l := time.Local
	time.Local = time.UTC
	defer func() { time.Local = l }()

	time20000101 := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
	timeNow := time.Now().Round(time.Second)
	timeNowPlusHour := timeNow.Add(time.Hour)
	timeNowMinusHour := timeNow.Add(-1 * time.Hour)
	timeNowMinusHourMinusExpiryDefaultDuration := timeNow.Add(-1 * time.Hour).Add(-1 * ExpiryDefaultDuration)

	testCases := []struct {
		name        string
		mni         *MNotificationInd
		wantExpire  time.Time
		wantExpired bool
	}{
		{},
		{"empty-empty",
			&MNotificationInd{},
			time.Time{}, false},
		{"empty-20000101",
			&MNotificationInd{Expiry: time20000101},
			time20000101, true},
		{"20000101-empty",
			&MNotificationInd{Received: time20000101},
			time20000101.Add(ExpiryDefaultDuration), true}, // Expired won't be true on very big default duration.
		{"20000101-20000101",
			&MNotificationInd{Received: time20000101, Expiry: time20000101},
			time20000101, true},
		{"empty-nowPlusHour",
			&MNotificationInd{Expiry: timeNowPlusHour},
			timeNowPlusHour, false},
		{"nowPlusHour-empty",
			&MNotificationInd{Received: timeNowPlusHour},
			timeNowPlusHour.Add(ExpiryDefaultDuration), false},
		{"nowPlusHour-nowPlusHour",
			&MNotificationInd{Received: timeNowPlusHour, Expiry: timeNowPlusHour},
			timeNowPlusHour, false},
		{"empty-nowMinusHour",
			&MNotificationInd{Expiry: timeNowMinusHour},
			timeNowMinusHour, true},
		{"nowMinusHour-empty",
			&MNotificationInd{Received: timeNowMinusHour},
			timeNowMinusHour.Add(ExpiryDefaultDuration), false}, // Expired won't be false on short default duration.
		{"nowMinusHourMinusExpiryDefaultDuration-empty",
			&MNotificationInd{Received: timeNowMinusHourMinusExpiryDefaultDuration},
			timeNowMinusHourMinusExpiryDefaultDuration.Add(ExpiryDefaultDuration), true}, // Expired won't be false on short default duration.
		{"nowMinusHour-nowMinusHour",
			&MNotificationInd{Received: timeNowMinusHour, Expiry: timeNowMinusHour},
			timeNowMinusHour, true},
		{"nowPlusHour-nowMinusHour",
			&MNotificationInd{Received: timeNowPlusHour, Expiry: timeNowMinusHour},
			timeNowMinusHour, true},
		{"nowMinusHour-nowPlusHour",
			&MNotificationInd{Received: timeNowMinusHour, Expiry: timeNowPlusHour},
			timeNowPlusHour, false},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.mni.Expire().Equal(tc.wantExpire) {
				t.Errorf("%#v.Expire() = %v, want %v", tc.mni, tc.mni.Expire(), tc.wantExpire)
			}
			if tc.mni.Expired() != tc.wantExpired {
				t.Errorf("%#v.Expired() = %v, want %v", tc.mni, tc.mni.Expired(), tc.wantExpired)
			}

		})
	}
}

func TestMNotificationInd_PopDebugError(t *testing.T) {
	debugUrl := "http://localhost:9191/mms"
	nodebugUrl := "http://123.456.789.012:3456/mms"
	testCases := []struct {
		name            string
		rawQuery, param string
		wantQuery       string
		wantError       error
	}{
		{},
		{
			"emptyParam",
			"", "",
			"", nil,
		},
		{
			"emptyParam-notFoundParam",
			"", "param",
			"", nil,
		},
		{
			"notFoundParam",
			"param=value", "notfound",
			"param=value", nil,
		},
		{
			"foundUnknownParam-invalidValue",
			"param=value", "param",
			"", nil,
		},
		{
			"foundUnknownParam-validZeroValue",
			"param=0", "param",
			"", nil,
		},
		{
			"foundUnknownParam-validOneValue",
			"param=1", "param",
			"", ForcedDebugError("param"),
		},
		{
			"foundUnknownParam-validFiveValue",
			"param=5", "param",
			"param=4", ForcedDebugError("param"),
		},
		{
			"notFoundParam",
			"param=value&other=1", "notfound",
			"param=value&other=1", nil,
		},
		{
			"foundUnknownParam-invalidValue",
			"param=value&other=1", "param",
			"other=1", nil,
		},
		{
			"foundUnknownParam-validZeroValue",
			"param=0&other=1", "param",
			"other=1", nil,
		},
		{
			"foundUnknownParam-validOneValue",
			"param=1&other=1", "param",
			"other=1", ForcedDebugError("param"),
		},
		{
			"foundUnknownParam-validFiveValue",
			"param=5&other=1", "param",
			"param=4&other=1", ForcedDebugError("param"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := url.ParseQuery(tc.rawQuery)
			if err != nil {
				t.Fatalf("Can't parse query \"%s\": %v", tc.rawQuery, err)
			}
			cl := debugUrl + "?" + tc.rawQuery
			mni := &MNotificationInd{ContentLocation: cl}
			err = mni.PopDebugError(tc.param)

			if !reflect.DeepEqual(err, tc.wantError) {
				t.Errorf("&MNotificationInd{ContentLocation: \"%s\"}.PopDebugError(\"%s\") = %v. want %v", cl, tc.param, err, tc.wantError)
			}

			uri, err := url.ParseRequestURI(mni.ContentLocation)
			if err != nil {
				t.Fatalf("Can't parse uri \"%s\": %v", tc.rawQuery, err)
			}
			wantValues, err := url.ParseQuery(tc.wantQuery)
			if err != nil {
				t.Fatalf("Can't parse wanted query \"%s\": %v", tc.wantQuery, err)
			}

			if !reflect.DeepEqual(wantValues, uri.Query()) {
				t.Errorf("MNotificationInd's ContentLocation (\"%s\") query after PopDebugError(...) is %v. want %v", cl, uri.Query(), wantValues)
			}

			// No debug url.
			cl = nodebugUrl + "?" + tc.rawQuery
			mni = &MNotificationInd{ContentLocation: cl}
			err = mni.PopDebugError(tc.param)

			if !reflect.DeepEqual(err, nil) {
				t.Errorf("&MNotificationInd{ContentLocation: \"%s\"}.PopDebugError(\"%s\") = %v. want %v", cl, tc.param, err, nil)
			}

			if cl != mni.ContentLocation {
				t.Errorf("MNotificationInd's ContentLocation (\"%s\") after PopDebugError(...) is %v. want %v", cl, mni.ContentLocation, cl)
			}

		})
	}
}
