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

// MMS Field names from OMA-WAP-MMS section 7.3
const (
	BCC                           = 0x01
	CC                            = 0x02
	X_MMS_CONTENT_LOCATION        = 0x03
	CONTENT_TYPE                  = 0x04
	DATE                          = 0x05
	X_MMS_DELIVERY_REPORT         = 0x06
	X_MMS_DELIVERY_TIME           = 0x07
	X_MMS_EXPIRY                  = 0x08
	FROM                          = 0x09
	X_MMS_MESSAGE_CLASS           = 0x0A
	MESSAGE_ID                    = 0x0B
	X_MMS_MESSAGE_TYPE            = 0x0C
	X_MMS_MMS_VERSION             = 0x0D
	X_MMS_MESSAGE_SIZE            = 0x0E
	X_MMS_PRIORITY                = 0x0F
	X_MMS_READ_REPORT             = 0x10
	X_MMS_REPORT_ALLOWED          = 0x11
	X_MMS_RESPONSE_STATUS         = 0x12
	X_MMS_RESPONSE_TEXT           = 0x13
	X_MMS_SENDER_VISIBILITY       = 0x14
	X_MMS_STATUS                  = 0x15
	SUBJECT                       = 0x16
	TO                            = 0x17
	X_MMS_TRANSACTION_ID          = 0x18
	X_MMS_RETRIEVE_STATUS         = 0x19
	X_MMS_RETRIEVE_TEXT           = 0x1A
	X_MMS_READ_STATUS             = 0x1B
	X_MMS_REPLY_CHARGING          = 0x1C
	X_MMS_REPLY_CHARGING_DEADLINE = 0x1D
	X_MMS_REPLY_CHARGING_ID       = 0x1E
	X_MMS_REPLY_CHARGING_SIZE     = 0x1F
	X_MMS_PREVIOUSLY_SENT_BY      = 0x20
	X_MMS_PREVIOUSLY_SENT_DATE    = 0x21
)

const (
	TYPE_SEND_REQ         = 0x80
	TYPE_SEND_CONF        = 0x81
	TYPE_NOTIFICATION_IND = 0x82
	TYPE_NOTIFYRESP_IND   = 0x83
	TYPE_RETRIEVE_CONF    = 0x84
	TYPE_ACKNOWLEDGE_IND  = 0x85
	TYPE_DELIVERY_IND     = 0x86
)

const (
	MMS_MESSAGE_VERSION_1_0 = 0x90
	MMS_MESSAGE_VERSION_1_1 = 0x91
	MMS_MESSAGE_VERSION_1_2 = 0x92
	MMS_MESSAGE_VERSION_1_3 = 0x93
)

// Date tokens defined in OMA-WAP-MMS section 7.2.10
const (
	TOKEN_DATE_ABS = 0x80
	TOKEN_DATE_REL = 0x81
)

// From tokens defined in OMA-WAP-MMS section 7.2.11
const (
	TOKEN_ADDRESS_PRESENT = 0x80
	TOKEN_INSERT_ADDRESS  = 0x81
)

// Message classes defined in OMA-WAP-MMS section 7.2.14
const (
	CLASS_PERSONAL      = 0x80
	CLASS_ADVERTISEMENT = 0x81
	CLASS_INFORMATIONAL = 0x82
	CLASS_AUTO          = 0x83
)

// MNotification in holds a m-notification.ind message defined in
// OMA-WAP-MMS-ENC section 6.2
type MNotificationInd struct {
	MMSReader
	Type, Version, Class, ReplyCharging, ReplyChargingDeadline, DeliveryReport byte
	TransactionId, From, Subject, ContentLocation, ReplyChargingId             string
	Expiry, Size                                                               uint64
}

// MNotification in holds a m-notifyresp.ind message defined in
// OMA-WAP-MMS-ENC-v1.1 section 6.2
type MNotifyRespInd struct {
	Type, Version, Status  byte
	TransactionId, Subject string
	ReportAllowed          bool
}

// MRetrieveConf in holds a m-retrieve.conf message defined in
// OMA-WAP-MMS-ENC-v1.1 section 6.3
type MRetrieveConf struct {
	MMSReader
	Type, Version, Status, Class, Priority, DeliveryReport, ReplyCharging, ReplyChargingDeadline, ReadReport, RetrieveStatus byte
	TransactionId, MessageId, From, To, Cc, Subject, ReplyChargingId, RetrieveText, FilePath                                 string
	ReportAllowed                                                                                                            bool
	Date                                                                                                                     uint64
	ContentType                                                                                                              ContentType
	DataParts                                                                                                                []ContentType
	Data                                                                                                                     []byte
}

type MMSReader interface{}

type MMSDecoder struct {
	Data   []byte
	Offset int
}

func NewMNotificationInd() *MNotificationInd {
	return &MNotificationInd{Type: TYPE_NOTIFICATION_IND}
}

func NewMRetrieveConf(filePath string) *MRetrieveConf {
	return &MRetrieveConf{Type: TYPE_RETRIEVE_CONF, FilePath: filePath}
}

func NewDecoder(data []byte) *MMSDecoder {
	return &MMSDecoder{Data: data}
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

func (dec *MMSDecoder) ReadString(reflectedPdu *reflect.Value, hdr string) (string, error) {
	dec.Offset++
	if dec.Data[dec.Offset] == 34 { // Skip the quote char(34) == "
		dec.Offset++
	}
	begin := dec.Offset
	//TODO protect this
	for ; dec.Data[dec.Offset] != 0; dec.Offset++ {
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
//or just decodes and discards if it's application specific, if the latter is the
//case it also returns false
func (dec *MMSDecoder) getParam() (byte, bool, error) {
	if dec.Data[dec.Offset]&0x80 != 0 {
		return dec.Data[dec.Offset] & 0x7f, true, nil
	} else {
		var param, value string
		var err error
		dec.Offset--
		if param, err = dec.ReadString(nil, ""); err != nil {
			return 0, false, err
		}
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
			ctMember := reflectedPdu.FieldByName("ContentType")
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
