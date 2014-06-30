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

	. "launchpad.net/gocheck"
)

type EncodeDecodeTestSuite struct {
	bytes *bytes.Buffer
	enc   *MMSEncoder
	dec   *MMSDecoder
}

var _ = Suite(&EncodeDecodeTestSuite{})

func (s *EncodeDecodeTestSuite) SetUpTest(c *C) {
	s.bytes = new(bytes.Buffer)
	s.enc = NewEncoder(s.bytes)
	c.Assert(s.enc.writeByte(0), IsNil)
}

func (s *EncodeDecodeTestSuite) TestString(c *C) {
	testStr := "'Hello World!"
	c.Assert(s.enc.writeString(testStr), IsNil)
	s.dec = NewDecoder(s.bytes.Bytes())

	str, err := s.dec.ReadString(nil, "")
	c.Assert(err, IsNil)
	c.Assert(str, Equals, testStr)
}

func (s *EncodeDecodeTestSuite) TestByte(c *C) {
	testBytes := []byte{0, 0x79, 0x80, 0x81}
	for i := range testBytes {
		c.Assert(s.enc.writeByte(testBytes[i]), IsNil)
	}
	bytes := s.bytes.Bytes()
	s.dec = NewDecoder(bytes)
	for i := range testBytes {
		b, err := s.dec.ReadByte(nil, "")
		c.Assert(err, IsNil)
		c.Assert(b, Equals, testBytes[i], Commentf("From testBytes[%d] and encoded bytes: %#x", i, bytes))
	}
}

func (s *EncodeDecodeTestSuite) TestInteger(c *C) {
	// 128 bounds short and long integers
	testInts := []uint64{512, 100, 127, 128, 129, 255, 256, 511, 3000}
	for i := range testInts {
		c.Assert(s.enc.writeInteger(testInts[i]), IsNil)
	}
	bytes := s.bytes.Bytes()
	s.dec = NewDecoder(bytes)
	for i := range testInts {
		integer, err := s.dec.ReadInteger(nil, "")
		c.Assert(err, IsNil)
		c.Check(integer, Equals, testInts[i], Commentf("%d != %d with encoded bytes starting at %d: %d", integer, testInts[i], i, bytes))
	}
}
