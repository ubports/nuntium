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
	"fmt"
	"reflect"
)

func NewDecoder(data []byte) *MMSDecoder {
	return &MMSDecoder{Data: data}
}

type MMSDecoder struct {
	Data   []byte
	Offset int
}

func (dec *MMSDecoder) ReadEncodedString(reflectedPdu *reflect.Value, hdr string) (string, error) {
	var length uint64
	var err error
	switch {
	case dec.Data[dec.Offset+1] < SHORT_LENGTH_MAX:
		var l byte
		l, err = dec.ReadShortInteger(nil, "")
		length = uint64(l)
	case dec.Data[dec.Offset+1] == LENGTH_QUOTE:
		dec.Offset++
		length, err = dec.ReadUintVar(nil, "")
	}
	if err != nil {
		return "", err
	}
	if length != 0 {
		charset, err := dec.ReadCharset(nil, "")
		if err != nil {
			return "", err
		}
		fmt.Println("Next string encoded with:", charset)
	}
	var str string
	if str, err = dec.ReadString(reflectedPdu, hdr); err != nil {
		return "", err
	}
	return str, nil
}

func (dec *MMSDecoder) ReadQ(reflectedPdu *reflect.Value) error {
	v, err := dec.ReadUintVar(nil, "")
	if err != nil {
		return err
	}
	q := float64(v)
	if q > 100 {
		q = (q - 100) / 1000
	} else {
		q = (q - 1) / 100
	}
	reflectedPdu.FieldByName("Q").SetFloat(q)
	return nil
}

func (dec *MMSDecoder) ReadLength(reflectedPdu *reflect.Value) (length uint64, err error) {
	switch {
	case dec.Data[dec.Offset+1] < SHORT_LENGTH_MAX:
		l, err := dec.ReadShortInteger(nil, "")
		v := uint64(l)
		reflectedPdu.FieldByName("Length").SetUint(v)
		return v, err
	case dec.Data[dec.Offset+1] == LENGTH_QUOTE:
		dec.Offset++
		l, err := dec.ReadUintVar(reflectedPdu, "Length")
		return l, err
	}
	return 0, fmt.Errorf("Unhandled length %#x @%d", dec.Data[dec.Offset+1], dec.Offset)
}

func (dec *MMSDecoder) ReadCharset(reflectedPdu *reflect.Value, hdr string) (string, error) {
	var charset string

	if dec.Data[dec.Offset] == ANY_CHARSET {
		dec.Offset++
		charset = "*"
	} else {
		charCode, err := dec.ReadInteger(nil, "")
		if err != nil {
			return "", err
		}
		var ok bool
		if charset, ok = CHARSETS[charCode]; !ok {
			return "", fmt.Errorf("Cannot find matching charset for %#x == %d", charCode, charCode)
		}
	}
	if hdr != "" {
		reflectedPdu.FieldByName("Charset").SetString(charset)
	}
	return charset, nil
}

func (dec *MMSDecoder) ReadMediaType(reflectedPdu *reflect.Value) (err error) {
	var mediaType string
	origOffset := dec.Offset

	if mt, err := dec.ReadInteger(nil, ""); err == nil && len(CONTENT_TYPES) > int(mt) {
		mediaType = CONTENT_TYPES[mt]
	} else {
		err = nil
		dec.Offset = origOffset
		mediaType, err = dec.ReadString(nil, "")
		if err != nil {
			return err
		}
	}

	reflectedPdu.FieldByName("MediaType").SetString(mediaType)
	fmt.Println("Media Type:", mediaType)
	return nil
}

func (dec *MMSDecoder) ReadString(reflectedPdu *reflect.Value, hdr string) (string, error) {
	dec.Offset++
	if dec.Data[dec.Offset] == 34 { // Skip the quote char(34) == "
		dec.Offset++
	}
	begin := dec.Offset
	for ; len(dec.Data) > dec.Offset; dec.Offset++ {
		if dec.Data[dec.Offset] == 0 {
			break
		}
	}
	if len(dec.Data) == dec.Offset {
		return "", fmt.Errorf("reached end of data while trying to read string: %s", dec.Data[begin:])
	}
	v := string(dec.Data[begin:dec.Offset])
	if hdr != "" {
		reflectedPdu.FieldByName(hdr).SetString(v)
		fmt.Printf("Setting %s to %s\n", hdr, v)
	}
	return v, nil
}

func (dec *MMSDecoder) ReadShortInteger(reflectedPdu *reflect.Value, hdr string) (byte, error) {
	dec.Offset++
	/*
		TODO fix use of short when not short
		if dec.Data[dec.Offset] & 0x80 == 0 {
			return 0, fmt.Errorf("Data on offset %d with value %#x is not a short integer", dec.Offset, dec.Data[dec.Offset])
		}
	*/
	v := dec.Data[dec.Offset] & 0x7F
	if hdr != "" {
		reflectedPdu.FieldByName(hdr).SetUint(uint64(v))
		fmt.Printf("Setting %s to %#x == %d\n", hdr, v, v)
	}
	return v, nil
}

func (dec *MMSDecoder) ReadByte(reflectedPdu *reflect.Value, hdr string) (byte, error) {
	dec.Offset++
	v := dec.Data[dec.Offset]
	if hdr != "" {
		reflectedPdu.FieldByName(hdr).SetUint(uint64(v))
		fmt.Printf("Setting %s to %#x == %d\n", hdr, v, v)
	}
	return v, nil
}

func (dec *MMSDecoder) ReadBytes(reflectedPdu *reflect.Value, hdr string) ([]byte, error) {
	dec.Offset++
	v := []byte(dec.Data[dec.Offset:])
	if hdr != "" {
		reflectedPdu.FieldByName(hdr).SetBytes(v)
		fmt.Printf("Setting %s to %#x == %d\n", hdr, v, v)
	}
	return v, nil
}

func (dec *MMSDecoder) ReadBoundedBytes(reflectedPdu *reflect.Value, hdr string, end int) ([]byte, error) {
	v := []byte(dec.Data[dec.Offset:end])
	if hdr != "" {
		reflectedPdu.FieldByName(hdr).SetBytes(v)
	}
	dec.Offset = end - 1
	return v, nil
}

// A UintVar is a variable lenght uint of up to 5 octects long where
// more octects available are indicated with the most significant bit
// set to 1
func (dec *MMSDecoder) ReadUintVar(reflectedPdu *reflect.Value, hdr string) (value uint64, err error) {
	dec.Offset++
	for dec.Data[dec.Offset]>>7 == 0x01 {
		value = value << 7
		value |= uint64(dec.Data[dec.Offset] & 0x7F)
		dec.Offset++
	}

	value = value << 7
	value |= uint64(dec.Data[dec.Offset] & 0x7F)
	if hdr != "" {
		reflectedPdu.FieldByName(hdr).SetUint(value)
		fmt.Printf("Setting %s to %d\n", hdr, value)
	}
	return value, nil
}

func (dec *MMSDecoder) ReadInteger(reflectedPdu *reflect.Value, hdr string) (uint64, error) {
	param := dec.Data[dec.Offset+1]
	var v uint64
	var err error
	switch {
	case param&0x80 != 0:
		var vv byte
		vv, err = dec.ReadShortInteger(nil, "")
		v = uint64(vv)
	default:
		v, err = dec.ReadLongInteger(nil, "")
	}
	if hdr != "" {
		reflectedPdu.FieldByName(hdr).SetUint(v)
		fmt.Printf("Setting %s to %d\n", hdr, v)
	}
	return v, err
}

func (dec *MMSDecoder) ReadLongInteger(reflectedPdu *reflect.Value, hdr string) (uint64, error) {
	dec.Offset++
	size := int(dec.Data[dec.Offset])
	dec.Offset++
	var v uint64
	endOffset := dec.Offset + size - 1
	v = v << 8
	for ; dec.Offset < endOffset; dec.Offset++ {
		v |= uint64(dec.Data[dec.Offset])
		v = v << 8
	}
	if hdr != "" {
		reflectedPdu.FieldByName(hdr).SetUint(uint64(v))
		fmt.Printf("Setting %s to %d\n", hdr, v)
	}
	return v, nil
}

//getParam reads the next parameter to decode and returns it if it's well known
//or just decodes and discards if it's application specific, if the latter is
//the case it also returns false
func (dec *MMSDecoder) getParam() (byte, bool, error) {
	if dec.Data[dec.Offset]&0x80 != 0 {
		return dec.Data[dec.Offset] & 0x7f, true, nil
	} else {
		var param, value string
		var err error
		dec.Offset--
		//Read the parameter name
		if param, err = dec.ReadString(nil, ""); err != nil {
			return 0, false, err
		}
		//Read the parameter value
		if value, err = dec.ReadString(nil, ""); err != nil {
			return 0, false, err
		}
		fmt.Println("Ignoring application header:", param, ":", value)
		return 0, false, nil
	}
}

func (dec *MMSDecoder) Decode(pdu MMSReader) (err error) {
	reflectedPdu := reflect.ValueOf(pdu).Elem()
	moreHdrToRead := true
	//fmt.Printf("len data: %d, data: %x\n", len(dec.Data), dec.Data)
	for ; (dec.Offset < len(dec.Data)) && moreHdrToRead; dec.Offset++ {
		//fmt.Printf("offset %d, value: %x\n", dec.Offset, dec.Data[dec.Offset])
		err = nil
		param, needsDecoding, err := dec.getParam()
		if err != nil {
			return err
		} else if !needsDecoding {
			continue
		}
		switch param {
		case X_MMS_MESSAGE_TYPE:
			dec.Offset++
			expectedType := byte(reflectedPdu.FieldByName("Type").Uint())
			parsedType := dec.Data[dec.Offset]
			//Unknown message types will be discarded. OMA-WAP-MMS-ENC-v1.1 section 7.2.16
			if parsedType != expectedType {
				err = fmt.Errorf("Expected message type %x got %x", expectedType, parsedType)
			}
		case FROM:
			dec.Offset++
			size := int(dec.Data[dec.Offset])
			dec.Offset++
			token := dec.Data[dec.Offset]
			switch token {
			case TOKEN_INSERT_ADDRESS:
				break
			case TOKEN_ADDRESS_PRESENT:
				// TODO add check for /TYPE=PLMN
				var from string
				from, err = dec.ReadString(&reflectedPdu, "From")
				// size - 2 == size - token - '0'
				if len(from) != size-2 {
					err = fmt.Errorf("From field is %d but expected size is %d", len(from), size-2)
				}
			default:
				err = fmt.Errorf("Unhandled token address in from field %x", token)
			}
		case X_MMS_EXPIRY:
			dec.Offset++
			size := int(dec.Data[dec.Offset])
			dec.Offset++
			token := dec.Data[dec.Offset]
			dec.Offset++
			var val uint
			endOffset := dec.Offset + size - 2
			for ; dec.Offset < endOffset; dec.Offset++ {
				val = (val << 8) | uint(dec.Data[dec.Offset])
			}
			// TODO add switch case for token
			fmt.Printf("Expiry token: %x\n", token)
			reflectedPdu.FieldByName("Expiry").SetUint(uint64(val))
			fmt.Printf("Message Expiry %d, %x\n", val, dec.Data[dec.Offset])
		case X_MMS_TRANSACTION_ID:
			_, err = dec.ReadString(&reflectedPdu, "TransactionId")
		case CONTENT_TYPE:
			ctMember := reflectedPdu.FieldByName("Content")
			if err = dec.ReadContentType(&ctMember); err != nil {
				return err
			}
			//application/vnd.wap.multipart.related and others
			if ctMember.FieldByName("MediaType").String() != "text/plain" {
				err = dec.ReadContentTypeParts(&reflectedPdu)
			} else {
				dec.Offset++
				_, err = dec.ReadBoundedBytes(&reflectedPdu, "Data", len(dec.Data))
			}
			moreHdrToRead = false
		case X_MMS_CONTENT_LOCATION:
			_, err = dec.ReadString(&reflectedPdu, "ContentLocation")
			moreHdrToRead = false
		case MESSAGE_ID:
			_, err = dec.ReadString(&reflectedPdu, "MessageId")
		case SUBJECT:
			_, err = dec.ReadEncodedString(&reflectedPdu, "Subject")
		case TO:
			_, err = dec.ReadEncodedString(&reflectedPdu, "To")
		case CC:
			_, err = dec.ReadEncodedString(&reflectedPdu, "Cc")
		case X_MMS_REPLY_CHARGING_ID:
			_, err = dec.ReadString(&reflectedPdu, "ReplyChargingId")
		case X_MMS_RETRIEVE_TEXT:
			_, err = dec.ReadString(&reflectedPdu, "RetrieveText")
		case X_MMS_MMS_VERSION:
			_, err = dec.ReadShortInteger(&reflectedPdu, "Version")
		case X_MMS_MESSAGE_CLASS:
			//TODO implement Token text form
			_, err = dec.ReadByte(&reflectedPdu, "Class")
		case X_MMS_REPLY_CHARGING:
			_, err = dec.ReadByte(&reflectedPdu, "ReplyCharging")
		case X_MMS_REPLY_CHARGING_DEADLINE:
			_, err = dec.ReadByte(&reflectedPdu, "ReplyChargingDeadLine")
		case X_MMS_PRIORITY:
			_, err = dec.ReadByte(&reflectedPdu, "Priority")
		case X_MMS_RETRIEVE_STATUS:
			_, err = dec.ReadByte(&reflectedPdu, "RetrieveStatus")
		case X_MMS_DELIVERY_REPORT:
			_, err = dec.ReadByte(&reflectedPdu, "DeliveryReport")
		case X_MMS_READ_REPORT:
			_, err = dec.ReadByte(&reflectedPdu, "ReadReport")
		case X_MMS_MESSAGE_SIZE:
			_, err = dec.ReadLongInteger(&reflectedPdu, "Size")
		case DATE:
			_, err = dec.ReadLongInteger(&reflectedPdu, "Date")
		default:
			fmt.Printf("Unhandled byte: %#0x\tdec: %d\tdec.Offset: %d\n", param, param, dec.Offset)
			return fmt.Errorf("Unhandled byte: %#0x\tdec: %d\tdec.Offset: %d\n", param, param, dec.Offset)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
