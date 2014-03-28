/*
 * Copyright 2014 Canonical Ltd.
 *
 * Authors:
 * Sergio Schvezov: sergio.schvezov@cannical.com
 *
 * This file is part of nuntium.
 *
 * nuntium is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; version 3.
 *
 * nuntium is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package ofono

import (
	"errors"
	"fmt"
	"reflect"
)

type PDU byte

const (
	CONNECT        PDU = 0x01
	CONNECT_REPLY  PDU = 0x02
	REDIRECT       PDU = 0x03
	REPLY          PDU = 0x04
	DISCONNECT     PDU = 0x05
	PUSH           PDU = 0x06
	CONFIRMED_PUSH PDU = 0x07
	SUSPEND        PDU = 0x08
	RESUME         PDU = 0x09
	GET            PDU = 0x40
	POST           PDU = 0x60
)

const (
	WSP_PARAMETER_TYPE_Q                  = 0x00
	WSP_PARAMETER_TYPE_CHARSET            = 0x01
	WSP_PARAMETER_TYPE_LEVEL              = 0x02
	WSP_PARAMETER_TYPE_TYPE               = 0x03
	WSP_PARAMETER_TYPE_NAME_DEFUNCT       = 0x05
	WSP_PARAMETER_TYPE_FILENAME_DEFUNCT   = 0x06
	WSP_PARAMETER_TYPE_DIFFERENCES        = 0x07
	WSP_PARAMETER_TYPE_PADDING            = 0x08
	WSP_PARAMETER_TYPE_CONTENT_TYPE       = 0x09
	WSP_PARAMETER_TYPE_START_DEFUNCT      = 0x0A
	WSP_PARAMETER_TYPE_START_INFO_DEFUNCT = 0x0B
	WSP_PARAMETER_TYPE_COMMENT_DEFUNCT    = 0x0C
	WSP_PARAMETER_TYPE_DOMAIN_DEFUNCT     = 0x0D
	WSP_PARAMETER_TYPE_MAX_AGE            = 0x0E
	WSP_PARAMETER_TYPE_PATH_DEFUNCT       = 0x0F
	WSP_PARAMETER_TYPE_SECURE             = 0x10
	WSP_PARAMETER_TYPE_SEC                = 0x11
	WSP_PARAMETER_TYPE_MAC                = 0x12
	WSP_PARAMETER_TYPE_CREATION_DATE      = 0x13
	WSP_PARAMETER_TYPE_MODIFICATION_DATE  = 0x14
	WSP_PARAMETER_TYPE_READ_DATE          = 0x15
	WSP_PARAMETER_TYPE_SIZE               = 0x16
	WSP_PARAMETER_TYPE_NAME               = 0x17
	WSP_PARAMETER_TYPE_FILENAME           = 0x18
	WSP_PARAMETER_TYPE_START              = 0x19
	WSP_PARAMETER_TYPE_START_INFO         = 0x1A
	WSP_PARAMETER_TYPE_COMMENT            = 0x1B
	WSP_PARAMETER_TYPE_DOMAIN             = 0x1C
	WSP_PARAMETER_TYPE_PATH               = 0x1D
	WSP_PARAMETER_TYPE_UNTYPED            = 0xFF
)

const (
	ACCEPT                = 0x00
	ACCEPT_CHARSET_1      = 0x01
	ACCEPT_ENCODING_1     = 0x02
	ACCEPT_LANGUAGE       = 0x03
	ACCEPT_RANGES         = 0x04
	AGE                   = 0x05
	ALLOW                 = 0x06
	AUTHORIZATION         = 0x07
	CACHE_CONTROL_1       = 0x08
	CONNECTION            = 0x09
	CONTENT_BASE          = 0x0A
	CONTENT_ENCODING      = 0x0B
	CONTENT_LANGUAGE      = 0x0C
	CONTENT_LENGTH        = 0x0D
	CONTENT_LOCATION      = 0x0E
	CONTENT_MD5           = 0x0F
	CONTENT_RANGE_1       = 0x10
	CONTENT_TYPE          = 0x11
	DATE                  = 0x12
	ETAG                  = 0x13
	EXPIRES               = 0x14
	FROM                  = 0x15
	HOST                  = 0x16
	IF_MODIFIED_SINCE     = 0x17
	IF_MATCH              = 0x18
	IF_NONE_MATCH         = 0x19
	IF_RANGE              = 0x1A
	IF_UNMODIFIED_SINCE   = 0x1B
	LOCATION              = 0x1C
	LAST_MODIFIED         = 0x1D
	MAX_FORWARDS          = 0x1E
	PRAGMA                = 0x1F
	PROXY_AUTHENTICATE    = 0x20
	PROXY_AUTHORIZATION   = 0x21
	PUBLIC                = 0x22
	RANGE                 = 0x23
	REFERER               = 0x24
	RETRY_AFTER           = 0x25
	SERVER                = 0x26
	TRANSFER_ENCODING     = 0x27
	UPGRADE               = 0x28
	USER_AGENT            = 0x29
	VARY                  = 0x2A
	VIA                   = 0x2B
	WARNING               = 0x2C
	WWW_AUTHENTICATE      = 0x2D
	CONTENT_DISPOSITION_1 = 0x2E
	X_WAP_APPLICATION_ID  = 0x2F
	X_WAP_CONTENT_URI     = 0x30
	X_WAP_INITIATOR_URI   = 0x31
	ACCEPT_APPLICATION    = 0x32
	BEARER_INDICATION     = 0x33
	PUSH_FLAG             = 0x34
	PROFILE               = 0x35
	PROFILE_DIFF          = 0x36
	PROFILE_WARNING_1     = 0x37
	EXPECT                = 0x38
	TE                    = 0x39
	TRAILER               = 0x3A
	ACCEPT_CHARSET        = 0x3B
	ACCEPT_ENCODING       = 0x3C
	CACHE_CONTROL_2       = 0x3D
	CONTENT_RANGE         = 0x3E
	X_WAP_TOD             = 0x3F
	CONTENT_ID            = 0x40
	SET_COOKIE            = 0x41
	COOKIE                = 0x42
	ENCODING_VERSION      = 0x43
	PROFILE_WARNING       = 0x44
	CONTENT_DISPOSITION   = 0x45
	X_WAP_SECURITY        = 0x46
	CACHE_CONTROL         = 0x47
)

type WSP interface {
	DecodeField() (interface{}, error)
}

type PushPDU struct {
	HeaderLength                             uint
	ContentLength                            int
	ApplicationId, EncodingVersion, PushFlag byte
	ContentType                              string
	Data                                     []byte
}

type PushPDUDecoder struct {
	data   []byte
	offset int
}

func NewDecoder(data []byte) *PushPDUDecoder {
	decoder := new(PushPDUDecoder)
	decoder.data = data
	return decoder
}

// The HeadersLen field specifies the length of the ContentType and Headers fields combined.
// The ContentType field contains the content type of the data. It conforms to the Content-Type value encoding specified
// in section 8.4.2.24, “Content type field”.
// The Headers field contains the push headers.
// The Data field contains the data pushed from the server. The length of the Data field is determined by the SDU size as
// provided to and reported from the underlying transport. The Data field starts immediately after the Headers field and
// ends at the end of the SDU.
func (dec *PushPDUDecoder) Decode(pdu *PushPDU) (err error) {
	if PDU(dec.data[1]) != PUSH {
		return errors.New(fmt.Sprintf("%x != %x is not a push PDU", PDU(dec.data[1]), PUSH))
	}
	// Move offset +tid +type = +2
	dec.offset = 2
	if pdu.HeaderLength, err = dec.decodeUintVar(); err != nil {
		return err
	}
	if pdu.ContentType, err = dec.decodeContentType(); err != nil {
		return err
	}
	if err = dec.decodeHeaders(pdu, int(pdu.HeaderLength)-(len(pdu.ContentType)+1)); err != nil {
		return err
	}
	pdu.Data = dec.data[(pdu.HeaderLength + 3):]
	return nil
}

// A UintVar is a variable lenght uint of up to 5 octects long where
// more octects available are indicated with the most significant bit
// set to 1
func (dec *PushPDUDecoder) decodeUintVar() (uint, error) {
	var val byte
	var n int

	for n = dec.offset; n < (dec.offset+5) && n < len(dec.data); n++ {
		val = (val << 7) | (dec.data[n] & 0x7f)
		if dec.data[n]&0x80 == 0 {
			break
		}
	}
	if dec.data[n]&0x80 == 1 {
		return 0, errors.New(fmt.Sprintf("Could not decode uintvar from %x", dec.data[n:]))
	}
	dec.offset = n + 1
	return uint(val), nil
}

func (dec *PushPDUDecoder) decodeContentType() (string, error) {
	var err error
	content, err := dec.DecodeField()
	if err != nil {
		return "", err
	}
	return reflect.ValueOf(content).String(), nil
}

func (dec *PushPDUDecoder) decodeHeaders(pdu *PushPDU, hdrLengthRemain int) error {
	var n int
	for n = dec.offset; n < (hdrLengthRemain + dec.offset); {
		param := dec.data[n] & 0x7F
		//fmt.Printf("byte: %#0x\tdec: %d\tn: %d\tdec.offset: %d\n", param, param, n, dec.offset)
		switch {
		case param == X_WAP_APPLICATION_ID:
			n++
			pdu.ApplicationId = dec.data[n] & 0x7F
			n++
		case param == PUSH_FLAG:
			n++
			pdu.PushFlag = dec.data[n] & 0x7F
			n++
		case param == ENCODING_VERSION:
			n++
			pdu.EncodingVersion = dec.data[n] & 0x7F
			n++
		case param == CONTENT_LENGTH:
			n++
			pdu.ContentLength = int(dec.data[n] & 0x7F)
			n++
		default:
			fmt.Printf("Unhandled byte: %#0x\tdec: %d\tn: %d\tdec.offset: %d\n", param, param, n, dec.offset)
			fmt.Println(pdu)
			return fmt.Errorf("Unhandled %x, %d, %d", param, n, dec.offset)
		}
	}
	dec.offset = n
	return nil
}

// decodeField decodes a header field following the following rules:
//
// WAP-230-WSP Section 8.4.1.2
// 0-30 octet is followed by the indicated number of octets
// 31 octet is followed by a uintvar
// 32-127 value is a nul terminated string
// 128-255 encoded 7bit value; no more data follows this octet
func (dec *PushPDUDecoder) DecodeField() (interface{}, error) {
	c := int(dec.data[dec.offset])
	switch {
	// Section 8.4.1.2 of WAP-230:
	// Field is encoded in text format
	case 32 < c && c < 127:
		var (
			text string
			n    int
			v    byte
		)
		for n, v = range dec.data[dec.offset:] {
			if v != 0 {
				continue
			}
			text = string(dec.data[dec.offset : dec.offset+n])
			break
		}
		dec.offset = dec.offset + n + 1
		return reflect.ValueOf(text).Interface(), nil
	}

	return nil, errors.New("Unhandled field type")
}
