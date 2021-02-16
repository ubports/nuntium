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
	"fmt"
	"reflect"
	"testing"

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

func TestMMSDecoder_ReadExpiry(t *testing.T) {
	testCases := []struct {
		name        string
		bytes       []byte
		offset      int
		destination interface{}
		wantExpiry  Expiry
		wantError   error
		wantOffset  int
		wantPanic   interface{}
	}{
		{
			"relative-5minutes",
			[]byte{0x88, 0x04, 0x81, 0x02, 0x01, 0x2c}, 0, &Expiry{},
			Expiry{ExpiryTokenRelative, 300}, nil, 5, nil,
		},
		{
			"relative-5minutes-offset",
			[]byte{0x00, 0x88, 0x04, 0x81, 0x02, 0x01, 0x2c}, 1, &Expiry{},
			Expiry{ExpiryTokenRelative, 300}, nil, 6, nil,
		},
		{
			"relative-1day",
			[]byte{0x88, 0x05, 0x81, 0x03, 0x01, 0x51, 0x80}, 0, &Expiry{},
			Expiry{ExpiryTokenRelative, 24 * 3600}, nil, 6, nil,
		},
		{
			"relative-2days",
			[]byte{0x88, 0x05, 0x81, 0x03, 0x02, 0xA3, 0x00}, 0, &Expiry{},
			Expiry{ExpiryTokenRelative, 2 * 24 * 3600}, nil, 6, nil,
		},
		{
			"absolute-date",
			[]byte{0x88, 0x06, 0x80, 0x04, 0x40, 0x19, 0xfe, 0x91}, 0, &Expiry{},
			Expiry{ExpiryTokenAbsolute, 1075445393}, nil, 7, nil,
		},
		{
			"error-expiry-length",
			[]byte{0x88, 0x05, 0x81, 0x02, 0x01, 0x2c}, 0, &Expiry{},
			Expiry{}, ErrorDecodeShortData{6, 6}, 1, nil,
		},
		{
			"error-value-length",
			[]byte{0x88, 0x04, 0x81, 0x03, 0x01, 0x2c}, 0, &Expiry{},
			Expiry{}, ErrorDecodeShortData{4, 6}, 1, "runtime error: index out of range [6] with length 6",
		},
		{
			"error-unknown-token",
			[]byte{0x88, 0x04, 0x82, 0x03, 0x01, 0x2c}, 0, &Expiry{},
			Expiry{0x82, 0}, ErrorDecodeUnknownExpiryToken(0x82), 2, nil,
		},
		{
			"error-inconsistent-offset-lower",
			[]byte{0x88, 0x05, 0x81, 0x02, 0x01, 0x2c, 0x00}, 0, &Expiry{},
			Expiry{ExpiryTokenRelative, 300}, ErrorDecodeInconsistentOffset{5, 6}, 5, nil,
		},
		{
			"error-inconsistent-offset-higher",
			[]byte{0x88, 0x03, 0x81, 0x02, 0x01, 0x2c, 0x00}, 0, &Expiry{},
			Expiry{ExpiryTokenRelative, 300}, ErrorDecodeInconsistentOffset{5, 4}, 5, nil,
		},
		{
			"relative-5minutes-nodestination",
			[]byte{0x88, 0x04, 0x81, 0x02, 0x01, 0x2c}, 0, nil,
			Expiry{ExpiryTokenRelative, 300}, nil, 5, nil,
		},
		{
			"error-destination-token",
			[]byte{0x88, 0x04, 0x81, 0x02, 0x01, 0x2c}, 0, &struct{}{},
			Expiry{ExpiryTokenRelative, 300}, nil, 5, nil,
		},
		{
			"error-destination-value",
			[]byte{0x88, 0x04, 0x81, 0x02, 0x01, 0x2c}, 0, &struct{ Token uint8 }{},
			Expiry{ExpiryTokenRelative, 300}, nil, 5, nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); !reflect.DeepEqual(r, tc.wantPanic) {
					_, isRuntimeError := r.(interface{ RuntimeError() })
					if isRuntimeError && fmt.Sprint(r) == fmt.Sprint(tc.wantPanic) {
						return
					}
					t.Errorf("Defered recover() = %#v, want %#v", r, tc.wantPanic)
				}
			}()

			dec := NewDecoder(tc.bytes)
			dec.Offset = tc.offset
			var reflectedPdu *reflect.Value
			if tc.destination != nil {
				elem := reflect.ValueOf(tc.destination).Elem()
				reflectedPdu = &elem
			}
			expiry, err := dec.ReadExpiry(reflectedPdu)

			if expiry != tc.wantExpiry || !reflect.DeepEqual(err, tc.wantError) {
				t.Errorf("MMSDecoder.ReadExpiry(%T) = (%v, %v), want (%v, %v)", tc.destination, expiry, err, tc.wantExpiry, tc.wantError)
			}
			if dec.Offset != tc.wantOffset {
				t.Errorf("After MMSDecoder.ReadExpiry(...), the MMSDecoder.Offset = %v, want %v", dec.Offset, tc.wantOffset)
			}
			if tc.destination != nil && err == nil {
				if reflectedPdu.FieldByName("Token").IsValid() && expiry.Token != uint8(reflectedPdu.FieldByName("Token").Uint()) {
					t.Errorf("Destination Token = %v, want %v", reflectedPdu.FieldByName("Token"), expiry.Token)
				}
				if reflectedPdu.FieldByName("Value").IsValid() && expiry.Value != uint64(reflectedPdu.FieldByName("Value").Uint()) {
					t.Errorf("Destination Value = %v, want %v", reflectedPdu.FieldByName("Value"), expiry.Value)
				}
			}
		})
	}
}
