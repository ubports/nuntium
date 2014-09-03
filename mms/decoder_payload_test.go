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

func (s *PayloadDecoderTestSuite) TestDecodeStringNoNullByteTerminator(c *C) {
	inputBytes, err := ioutil.ReadFile("test_payloads/m-send.conf_success")
	c.Assert(err, IsNil)

	mSendConf := NewMSendConf()
	dec := NewDecoder(inputBytes)
	err = dec.Decode(mSendConf)
	c.Assert(err, IsNil)
	c.Check(mSendConf.ResponseStatus, Equals, ResponseStatusOk)
	c.Check(mSendConf.TransactionId, Equals, "ad6babe2628710c443cdeb3ff39679ac")
}
