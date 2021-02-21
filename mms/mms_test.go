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
