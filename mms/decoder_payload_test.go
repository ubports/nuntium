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
	"io/ioutil"
	"reflect"
	"testing"
	"time"

	. "launchpad.net/gocheck"
)

type PayloadDecoderTestSuite struct{}

var _ = Suite(&PayloadDecoderTestSuite{})

func (s *PayloadDecoderTestSuite) TestDecodeSuccessfulMSendConf(c *C) {
	inputBytes, err := ioutil.ReadFile("test_payloads/m-send.conf_success")
	c.Assert(err, IsNil)

	mSendConf := NewMSendConf()
	dec := NewDecoder(inputBytes)
	err = dec.Decode(mSendConf)
	c.Assert(err, IsNil)
	c.Check(mSendConf.ResponseStatus, Equals, ResponseStatusOk)
	c.Check(mSendConf.TransactionId, Equals, "ad6babe2628710c443cdeb3ff39679ac")
}

func (s *PayloadDecoderTestSuite) TestDecodeSuccessfulMRetrieveConf(c *C) {
	inputBytes, err := ioutil.ReadFile("test_payloads/m-retrieve.conf_success")
	c.Assert(err, IsNil)

	mRetrieveConf := NewMRetrieveConf("55555555")
	dec := NewDecoder(inputBytes)
	err = dec.Decode(mRetrieveConf)
	c.Assert(err, IsNil)
	c.Check(mRetrieveConf.MessageId, Equals, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	c.Check(mRetrieveConf.From, Equals, "11111111111/TYPE=PLMN")
	c.Check(mRetrieveConf.To[0], Equals, "2222222222/TYPE=PLMN")
}

func (s *PayloadDecoderTestSuite) TestDecodeInvalidMSendConf(c *C) {
	inputBytes := []byte(`<html><head><title>719</title><meta http-equiv="Cache-Control" content="max-age=0" /><meta http-equiv="Cache-control" content="no-cache" /></head><body><h3 align="center">Disculpe,ha ocurrido un error: Failure to Query from Radius Server</h3><br/><p>Por favor, regrese al menu anterior o acceda al siguiente link.<br/></p><ul><li><a href="http://wap.personal.com.ar/"><strong>Home Personal</strong></a></li></ul></body></html>^M`)

	mSendConf := NewMSendConf()
	dec := NewDecoder(inputBytes)
	err := dec.Decode(mSendConf)
	c.Check(err, NotNil)
	c.Check(mSendConf.ResponseStatus, Equals, byte(0x0))
	c.Check(mSendConf.TransactionId, Equals, "")
	mSendConf.Status()
}

type testDecodeMNotificationInd_missingReceived struct {
	Version, Class  byte
	ContentLocation string
	From            string
	Expiry          time.Time
	Size            uint64
}
type testDecodeMNotificationInd_invalidReceived struct {
	Received        string
	Version, Class  byte
	ContentLocation string
	From            string
	Expiry          time.Time
	Size            uint64
}

func TestMMSDecoder_Decode_MNotificationInd(t *testing.T) {
	fileSuccess := "test_payloads/m-notification.ind_success"
	bytesSuccess, err := ioutil.ReadFile(fileSuccess)
	if err != nil {
		t.Fatalf("Can't load test data from %s due to error: %v", fileSuccess, err)
	}
	time20000101 := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)

	testCases := []struct {
		name      string
		pdu       interface{}
		bytes     []byte
		wantPdu   interface{}
		wantError error
	}{
		{"empty-success",
			&MNotificationInd{}, bytesSuccess,
			&MNotificationInd{
				Version:         MMS_MESSAGE_VERSION_1_0,
				From:            "+543515924906/TYPE=PLMN",
				Class:           ClassPersonal,
				Size:            29696,
				Expiry:          time.Time{}.Add(2*24*time.Hour - 1*time.Second),
				ContentLocation: "http://localhost:9191/mms",
			}, nil},
		{"20000101-success",
			&MNotificationInd{Received: time20000101}, bytesSuccess,
			&MNotificationInd{
				Received:        time20000101,
				Version:         MMS_MESSAGE_VERSION_1_0,
				From:            "+543515924906/TYPE=PLMN",
				Class:           ClassPersonal,
				Size:            29696,
				Expiry:          time20000101.Add(2*24*time.Hour - 1*time.Second),
				ContentLocation: "http://localhost:9191/mms",
			}, nil},
		{"missingReceived-success",
			&testDecodeMNotificationInd_missingReceived{}, bytesSuccess,
			&testDecodeMNotificationInd_missingReceived{
				Version:         MMS_MESSAGE_VERSION_1_0,
				From:            "+543515924906/TYPE=PLMN",
				Class:           ClassPersonal,
				Size:            29696,
				Expiry:          time.Time{}.Add(2*24*time.Hour - 1*time.Second),
				ContentLocation: "http://localhost:9191/mms",
			}, nil},
		{"invalidReceived-success",
			&testDecodeMNotificationInd_invalidReceived{}, bytesSuccess,
			&testDecodeMNotificationInd_invalidReceived{
				Version:         MMS_MESSAGE_VERSION_1_0,
				From:            "+543515924906/TYPE=PLMN",
				Class:           ClassPersonal,
				Size:            29696,
				Expiry:          time.Time{}.Add(2*24*time.Hour - 1*time.Second),
				ContentLocation: "http://localhost:9191/mms",
			}, nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Logf("%#v", tc.pdu)
			dec := NewDecoder(tc.bytes)
			err := dec.Decode(tc.pdu)
			t.Log(dec.GetLog())
			if err != tc.wantError {
				t.Errorf("MMSDecoder.Decode(%#v) = %v, want %v", tc.pdu, err, tc.wantError)
			}
			if !reflect.DeepEqual(tc.pdu, tc.wantPdu) {
				t.Errorf("After MMSDecoder.Decode(...), the param is \n\t%#v, want \n\t%#v", tc.pdu, tc.wantPdu)
			}
		})
	}
}
