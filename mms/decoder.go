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
	"log"
	"reflect"
)

func NewDecoder(data []byte) *MMSDecoder {
	return &MMSDecoder{Data: data}
}

type MMSDecoder struct {
	Data   []byte
	Offset int
	log    string
}

func (dec *MMSDecoder) setPduField(pdu *reflect.Value, name string, v interface{},
	setter func(*reflect.Value, interface{})) {

	if name != "" {
		field := pdu.FieldByName(name)
		if field.IsValid() {
			setter(&field, v)
			dec.log = dec.log + fmt.Sprintf("Setting %s to %v\n", name, v)
		} else {
			log.Println("Field", name, "not in decoding structure")
		}
	}
}

func setterString(field *reflect.Value, v interface{}) { field.SetString(v.(string)) }
func setterUint64(field *reflect.Value, v interface{}) { field.SetUint(v.(uint64)) }
func setterSlice(field *reflect.Value, v interface{})  { field.SetBytes(v.([]byte)) }

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
		dec.log = dec.log + fmt.Sprintf("Next string encoded with: %s\n", charset)
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

// ReadLength reads the length from the next position according to section
// 8.4.2.2 of WAP-230-WSP-20010705-a.
//
// Value-length = Short-length | (Length-quote Length)
// ; Value length is used to indicate the length of the value to follow
// Short-length = <Any octet 0-30> (0x7f to check for short)
// Length-quote = <Octet 31>
// Length = Uintvar-integer
func (dec *MMSDecoder) ReadLength(reflectedPdu *reflect.Value) (length uint64, err error) {
	switch {
	case dec.Data[dec.Offset+1]&0x7f <= SHORT_LENGTH_MAX:
		l, err := dec.ReadShortInteger(nil, "")
		v := uint64(l)
		if reflectedPdu != nil {
			reflectedPdu.FieldByName("Length").SetUint(v)
		}
		return v, err
	case dec.Data[dec.Offset+1] == LENGTH_QUOTE:
		dec.Offset++
		var hdr string
		if reflectedPdu != nil {
			hdr = "Length"
		}
		return dec.ReadUintVar(reflectedPdu, hdr)
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

func (dec *MMSDecoder) ReadMediaType(reflectedPdu *reflect.Value, hdr string) (err error) {
	var mediaType string
	var endOffset int
	origOffset := dec.Offset

	if dec.Data[dec.Offset+1] <= SHORT_LENGTH_MAX || dec.Data[dec.Offset+1] == LENGTH_QUOTE {
		if length, err := dec.ReadLength(nil); err != nil {
			return err
		} else {
			endOffset = int(length) + dec.Offset
		}
	}

	if dec.Data[dec.Offset+1] >= TEXT_MIN && dec.Data[dec.Offset+1] <= TEXT_MAX {
		if mediaType, err = dec.ReadString(nil, ""); err != nil {
			return err
		}
	} else if mt, err := dec.ReadInteger(nil, ""); err == nil && len(CONTENT_TYPES) > int(mt) {
		mediaType = CONTENT_TYPES[mt]
	} else {
		return fmt.Errorf("cannot decode media type for field beginning with %#x@%d", dec.Data[origOffset], origOffset)
	}

	// skip the rest of the content type params
	if endOffset > 0 {
		dec.Offset = endOffset
	}

	reflectedPdu.FieldByName(hdr).SetString(mediaType)
	dec.log = dec.log + fmt.Sprintf("%s: %s\n", hdr, mediaType)

	return nil
}

func (dec *MMSDecoder) ReadTo(reflectedPdu *reflect.Value) error {
	// field in the MMS protocol
	toField, err := dec.ReadEncodedString(reflectedPdu, "")
	if err != nil {
		return err
	}
	// field in the golang structure
	to := reflectedPdu.FieldByName("To")
	toSlice := reflect.Append(to, reflect.ValueOf(toField))
	reflectedPdu.FieldByName("To").Set(toSlice)
	return err
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
	dec.setPduField(reflectedPdu, hdr, v, setterString)

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
	dec.setPduField(reflectedPdu, hdr, uint64(v), setterUint64)

	return v, nil
}

func (dec *MMSDecoder) ReadByte(reflectedPdu *reflect.Value, hdr string) (byte, error) {
	dec.Offset++
	v := dec.Data[dec.Offset]
	dec.setPduField(reflectedPdu, hdr, uint64(v), setterUint64)

	return v, nil
}

func (dec *MMSDecoder) ReadBoundedBytes(reflectedPdu *reflect.Value, hdr string, end int) ([]byte, error) {
	v := []byte(dec.Data[dec.Offset:end])
	dec.setPduField(reflectedPdu, hdr, v, setterSlice)
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
	dec.setPduField(reflectedPdu, hdr, value, setterUint64)

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
	dec.setPduField(reflectedPdu, hdr, v, setterUint64)

	return v, err
}

func (dec *MMSDecoder) ReadLongInteger(reflectedPdu *reflect.Value, hdr string) (uint64, error) {
	dec.Offset++
	size := int(dec.Data[dec.Offset])
	if size > SHORT_LENGTH_MAX {
		//TODO:jezek Why is SHORT_LENGTH_MAX = 30, when later you're storing the size*bytes in an uint64 in which can fit only 8*bytes?
		return 0, fmt.Errorf("cannot encode long integer, lenght was %d but expected %d", size, SHORT_LENGTH_MAX)
	}
	dec.Offset++
	end := dec.Offset + size
	var v uint64
	for ; dec.Offset < end; dec.Offset++ {
		v = v << 8
		v |= uint64(dec.Data[dec.Offset])
	}
	dec.Offset--
	dec.setPduField(reflectedPdu, hdr, v, setterUint64)

	return v, nil
}

// ReadExpiry reads the expiry data from the next position according to OMA-TS-MMS_ENC-V1_3-20110913-A.
// If not nil reflectedPdu provided, it's elements "Token"(uint8) and "Value"(uint64) will be filled with decoded values. If the elements are not valid, only a log message will be written.
//
// 7.3.20 (X-Mms-Expiry Field) of
// Expiry-value = Value-length (Absolute-token Date-value | Relative-token Delta-seconds-value)
// Absolute-token = <Octet 128>
// Relative-token = <Octet 129>
//
// 7.3.12 (Date Field) of OMA-TS-MMS_ENC-V1_3-20110913-A
// Date-value = Long-integer
// In seconds from 1970-01-01, 00:00:00 GMT.
//
// 7.3.15 (Delta-Seconds-Value) of OMA-TS-MMS_ENC-V1_3-20110913-A
// Delta-seconds-value = Long-integer
func (dec *MMSDecoder) ReadExpiry(reflectedPdu *reflect.Value) (expiry Expiry, err error) {
	length, err := dec.ReadLength(nil)
	if err != nil {
		return expiry, err
	}
	endOffset := dec.Offset + int(length)
	if endOffset >= len(dec.Data) {
		return expiry, ErrorDecodeShortData{len(dec.Data), endOffset}
	}

	tokenField, valueField := "", ""
	if reflectedPdu != nil {
		tokenField, valueField = "Token", "Value"
	}

	expiry.Token, err = dec.ReadByte(reflectedPdu, tokenField)
	if err != nil {
		return expiry, err
	}
	if expiry.Token != ExpiryTokenAbsolute && expiry.Token != ExpiryTokenRelative {
		return expiry, ErrorDecodeUnknownExpiryToken(expiry.Token)
	}

	expiry.Value, err = dec.ReadLongInteger(reflectedPdu, valueField)
	if err != nil {
		return expiry, err
	}

	if dec.Offset != endOffset {
		return expiry, ErrorDecodeInconsistentOffset{dec.Offset, endOffset}
	}

	dec.log = dec.log + fmt.Sprintf("Message Expiry %#v\n", expiry)
	return expiry, nil
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
		dec.log = dec.log + fmt.Sprintf("Ignoring application header: %#x: %s", param, value)
		return 0, false, nil
	}
}

func (dec *MMSDecoder) skipFieldValue() error {
	switch {
	case dec.Data[dec.Offset+1] < LENGTH_QUOTE:
		l, err := dec.ReadByte(nil, "")
		if err != nil {
			return err
		}
		length := int(l)
		if dec.Offset+length >= len(dec.Data) {
			return fmt.Errorf("Bad field value length")
		}
		dec.Offset += length
		return nil
	case dec.Data[dec.Offset+1] == LENGTH_QUOTE:
		dec.Offset++
		// TODO These tests should be done in basic read functions
		if dec.Offset+1 >= len(dec.Data) {
			return fmt.Errorf("Bad uintvar")
		}
		l, err := dec.ReadUintVar(nil, "")
		if err != nil {
			return err
		}
		length := int(l)
		if dec.Offset+length >= len(dec.Data) {
			return fmt.Errorf("Bad field value length")
		}
		dec.Offset += length
		return nil
	case dec.Data[dec.Offset+1] <= TEXT_MAX:
		_, err := dec.ReadString(nil, "")
		return err
	}
	// case dec.Data[dec.Offset + 1] > TEXT_MAX
	_, err := dec.ReadShortInteger(nil, "")
	return err
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
			valStart := dec.Offset
			dec.Offset++
			token := dec.Data[dec.Offset]
			switch token {
			case TOKEN_INSERT_ADDRESS:
				break
			case TOKEN_ADDRESS_PRESENT:
				// TODO add check for /TYPE=PLMN
				_, err = dec.ReadEncodedString(&reflectedPdu, "From")
				if valStart+size != dec.Offset {
					err = fmt.Errorf("From field length is %d but expected size is %d",
						dec.Offset-valStart, size)
				}
			default:
				err = fmt.Errorf("Unhandled token address in from field %x", token)
			}
		case X_MMS_EXPIRY:
			var expiryPdu *reflect.Value
			if e := reflectedPdu.FieldByName("Expiry"); e.IsValid() {
				expiryPdu = &e
			} else {
				log.Println("Field Expiry not in decoding structure")
			}
			_, err = dec.ReadExpiry(expiryPdu)
		case X_MMS_TRANSACTION_ID:
			_, err = dec.ReadString(&reflectedPdu, "TransactionId")
		case CONTENT_TYPE:
			ctMember := reflectedPdu.FieldByName("Content")
			if err = dec.ReadAttachment(&ctMember); err != nil {
				return err
			}
			//application/vnd.wap.multipart.related and others
			if ctMember.FieldByName("MediaType").String() != "text/plain" {
				err = dec.ReadAttachmentParts(&reflectedPdu)
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
			err = dec.ReadTo(&reflectedPdu)
		case CC:
			_, err = dec.ReadEncodedString(&reflectedPdu, "Cc")
		case X_MMS_REPLY_CHARGING_ID:
			_, err = dec.ReadString(&reflectedPdu, "ReplyChargingId")
		case X_MMS_RETRIEVE_TEXT:
			_, err = dec.ReadString(&reflectedPdu, "RetrieveText")
		case X_MMS_MMS_VERSION:
			// TODO This should be ReadShortInteger instead, but we read it
			// as a byte because we are not properly encoding the version
			// either, as we are using the raw value there. To fix this we
			// need to change the encoder and the MMS_MESSAGE_VERSION_1_X
			// constants.
			_, err = dec.ReadByte(&reflectedPdu, "Version")
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
		case X_MMS_RESPONSE_STATUS:
			_, err = dec.ReadByte(&reflectedPdu, "ResponseStatus")
		case X_MMS_RESPONSE_TEXT:
			_, err = dec.ReadString(&reflectedPdu, "ResponseText")
		case X_MMS_DELIVERY_REPORT:
			_, err = dec.ReadByte(&reflectedPdu, "DeliveryReport")
		case X_MMS_READ_REPORT:
			_, err = dec.ReadByte(&reflectedPdu, "ReadReport")
		case X_MMS_MESSAGE_SIZE:
			_, err = dec.ReadLongInteger(&reflectedPdu, "Size")
		case DATE:
			_, err = dec.ReadLongInteger(&reflectedPdu, "Date")
		default:
			log.Printf("Skipping unrecognized header 0x%02x", param)
			err = dec.skipFieldValue()
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (dec *MMSDecoder) GetLog() string {
	return dec.log
}
