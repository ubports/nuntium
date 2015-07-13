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
