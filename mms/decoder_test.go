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
	"errors"

	. "launchpad.net/gocheck"
)

type DecoderTestSuite struct{}

var _ = Suite(&DecoderTestSuite{})

func (s *DecoderTestSuite) TestDecodeStringNoNullByteTerminator(c *C) {
	inputBytes := []byte{
		//stub byte
		0x80,
		//<html>
		0x3c, 0x68, 0x74, 0x6d, 0x6c, 0x3e,
	}
	expectedErr := errors.New("reached end of data while trying to read string: <html>")
	dec := NewDecoder(inputBytes)
	str, err := dec.ReadString(nil, "")
	c.Check(str, Equals, "")
	c.Check(err, DeepEquals, expectedErr)
}

func (s *DecoderTestSuite) TestDecodeStringWithNullByteTerminator(c *C) {
	inputBytes := []byte{
		//stub byte
		0x80,
		//<smil>
		0x3c, 0x73, 0x6d, 0x69, 0x6c, 0x3e, 0x00,
	}
	dec := NewDecoder(inputBytes)
	str, err := dec.ReadString(nil, "")
	c.Check(str, Equals, "<smil>")
	c.Check(err, IsNil)
}
