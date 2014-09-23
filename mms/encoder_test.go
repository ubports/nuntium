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
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	. "launchpad.net/gocheck"
)

type EncoderTestSuite struct{}

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

var _ = Suite(&EncoderTestSuite{})

func (s *EncoderTestSuite) TestEncodeMNotifyRespIndRetrievedWithReports(c *C) {
	expectedBytes := []byte{
		//Message Type m-notifyresp.ind
		0x8C, 0x83,
		// Transaction Id
		0x98, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x00,
		// MMS Version 1.3
		0x8D, 0x93,
		// Status retrieved
		0x95, 0x81,
		// Report Allowed No
		0x91, 0x81,
	}
	mNotifyRespInd := &MNotifyRespInd{
		UUID:          "1",
		Type:          TYPE_NOTIFYRESP_IND,
		TransactionId: "0123456",
		Version:       MMS_MESSAGE_VERSION_1_3,
		Status:        STATUS_RETRIEVED,
		ReportAllowed: ReportAllowedNo,
	}
	var outBytes bytes.Buffer
	enc := NewEncoder(&outBytes)
	c.Assert(enc.Encode(mNotifyRespInd), IsNil)
	c.Assert(outBytes.Bytes(), DeepEquals, expectedBytes)
}

func (s *EncoderTestSuite) TestEncodeMNotifyRespIndDeffered(c *C) {
	expectedBytes := []byte{
		//Message Type m-notifyresp.ind
		0x8C, 0x83,
		// Transaction Id
		0x98, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x00,
		// MMS Version 1.3
		0x8D, 0x93,
		// Status deffered
		0x95, 0x83,
		// Report Allowed No
		0x91, 0x81,
	}
	mNotifyRespInd := &MNotifyRespInd{
		UUID:          "1",
		Type:          TYPE_NOTIFYRESP_IND,
		TransactionId: "0123456",
		Version:       MMS_MESSAGE_VERSION_1_3,
		Status:        STATUS_DEFERRED,
		ReportAllowed: ReportAllowedNo,
	}
	var outBytes bytes.Buffer
	enc := NewEncoder(&outBytes)
	c.Assert(enc.Encode(mNotifyRespInd), IsNil)
	c.Assert(outBytes.Bytes(), DeepEquals, expectedBytes)
}

func (s *EncoderTestSuite) TestEncodeMNotifyRespIndRetrievedWithoutReports(c *C) {
	expectedBytes := []byte{
		//Message Type m-notifyresp.ind
		0x8C, 0x83,
		// Transaction Id
		0x98, 0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x00,
		// MMS Version 1.3
		0x8D, 0x93,
		// Status retrieved
		0x95, 0x81,
		// Report Allowed Yes
		0x91, 0x80,
	}
	mNotifyRespInd := &MNotifyRespInd{
		UUID:          "1",
		Type:          TYPE_NOTIFYRESP_IND,
		TransactionId: "0123456",
		Version:       MMS_MESSAGE_VERSION_1_3,
		Status:        STATUS_RETRIEVED,
		ReportAllowed: ReportAllowedYes,
	}
	var outBytes bytes.Buffer
	enc := NewEncoder(&outBytes)
	c.Assert(enc.Encode(mNotifyRespInd), IsNil)
	c.Assert(outBytes.Bytes(), DeepEquals, expectedBytes)
}

func (s *EncoderTestSuite) TestEncodeMSendReq(c *C) {
	tmp, err := ioutil.TempFile("", "")
	c.Assert(err, IsNil)
	tmp.Close()
	defer os.Remove(tmp.Name())
	err = ioutil.WriteFile(tmp.Name(), []byte{1, 2, 3, 4, 5, 6}, 0644)
	c.Assert(err, IsNil)

	att, err := NewAttachment("text0", "text0.txt", tmp.Name())
	c.Assert(err, IsNil)

	attachments := []*Attachment{att}

	recipients := []string{"+12345"}
	mSendReq := NewMSendReq(recipients, attachments, false)

	var outBytes bytes.Buffer
	enc := NewEncoder(&outBytes)
	err = enc.Encode(mSendReq)
	c.Assert(err, IsNil)
}
