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

	"github.com/ubuntu-phonedations/nuntium/mms"
)

type PDU byte

type PushPDU struct {
	HeaderLength                             uint64
	ContentLength                            uint64
	ApplicationId, EncodingVersion, PushFlag byte
	ContentType                              string
	Data                                     []byte
}

type PushPDUDecoder struct {
	mms.MMSDecoder
}

func NewDecoder(data []byte) *PushPDUDecoder {
	decoder := new(PushPDUDecoder)
	decoder.MMSDecoder.Data = data
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
	if PDU(dec.Data[1]) != PUSH {
		return errors.New(fmt.Sprintf("%x != %x is not a push PDU", PDU(dec.Data[1]), PUSH))
	}
	// Move offset +tid +type = +2
	dec.Offset = 1
	rValue := reflect.ValueOf(pdu).Elem()
	if _, err = dec.ReadUintVar(&rValue, "HeaderLength"); err != nil {
		return err
	}
	if err = dec.ReadMediaType(&rValue, "ContentType"); err != nil {
		return err
	}
	dec.Offset++
	remainHeaders := int(pdu.HeaderLength) - dec.Offset + 3
	if err = dec.decodeHeaders(pdu, remainHeaders); err != nil {
		return err
	}
	pdu.Data = dec.Data[(pdu.HeaderLength + 3):]
	return nil
}

func (dec *PushPDUDecoder) decodeHeaders(pdu *PushPDU, hdrLengthRemain int) error {
	rValue := reflect.ValueOf(pdu).Elem()
	var err error
	for ; dec.Offset < (hdrLengthRemain + dec.Offset); dec.Offset++ {
		param := dec.Data[dec.Offset] & 0x7F
		switch param {
		case X_WAP_APPLICATION_ID:
			_, err = dec.ReadInteger(&rValue, "ApplicationId")
		case PUSH_FLAG:
			_, err = dec.ReadShortInteger(&rValue, "PushFlag")
		case ENCODING_VERSION:
			dec.Offset++
			pdu.EncodingVersion = dec.Data[dec.Offset] & 0x7F
			dec.Offset++
		case CONTENT_LENGTH:
			_, err = dec.ReadInteger(&rValue, "ContentLength")
		case X_WAP_INITIATOR_URI:
			var v string
			v, err = dec.ReadString(nil, "")
			fmt.Println("Unsaved value decoded:", v)
		default:
			err = fmt.Errorf("Unhandled header data %#x @%d", dec.Data[dec.Offset], dec.Offset)
		}
		if err != nil {
			return fmt.Errorf("error while decoding %#x @%d: ", param, dec.Offset, err)
		} else if pdu.ApplicationId != 0 {
			return nil
		}

	}
	return nil
}
