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
	"io"
	"log"
	"reflect"
)

type MMSEncoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *MMSEncoder {
	return &MMSEncoder{w}
}

func (enc *MMSEncoder) Encode(pdu MMSWriter) error {
	rPdu := reflect.ValueOf(pdu).Elem()

	//The order of the following fields doens't matter much
	typeOfPdu := rPdu.Type()
	var err error
	for i := 0; i < rPdu.NumField(); i++ {
		fieldName := typeOfPdu.Field(i).Name
		encodeTag := typeOfPdu.Field(i).Tag.Get("encode")
		f := rPdu.Field(i)

		if encodeTag == "no" {
			continue
		}
		switch fieldName {
		case "Type":
			err = enc.writeByteParam(X_MMS_MESSAGE_TYPE, byte(f.Uint()))
		case "Version":
			err = enc.writeByteParam(X_MMS_MMS_VERSION, byte(f.Uint()))
		case "TransactionId":
			err = enc.writeStringParam(X_MMS_TRANSACTION_ID, f.String())
		case "Status":
			err = enc.writeByteParam(X_MMS_STATUS, byte(f.Uint()))
		case "ReportAllowed":
			err = enc.writeReportAllowedParam(f.Bool())
		default:
			if encodeTag == "optional" {
				log.Printf("Unhandled optional field %s", fieldName)
			} else {
				panic(fmt.Sprintf("missing encoding for mandatory field %s", fieldName))
			}
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (enc *MMSEncoder) setParam(param byte) error {
	return enc.writeByte(param | 0x80)
}

func (enc *MMSEncoder) writeStringParam(param byte, s string) error {
	if err := enc.setParam(param); err != nil {
		return err
	}
	return enc.writeString(s)
}

func (enc *MMSEncoder) writeByteParam(param byte, b byte) error {
	if err := enc.setParam(param); err != nil {
		return err
	}
	return enc.writeByte(b)
}

func (enc *MMSEncoder) writeReportAllowedParam(reportAllowed bool) error {
	if err := enc.setParam(X_MMS_REPORT_ALLOWED); err != nil {
		return err
	}
	var b byte
	if reportAllowed {
		b = REPORT_ALLOWED_YES
	} else {
		b = REPORT_ALLOWED_NO
	}
	return enc.writeByte(b)
}

func (enc *MMSEncoder) writeString(s string) error {
	bytes := []byte(s)
	bytes = append(bytes, 0)
	_, err := enc.w.Write(bytes)
	return err
}

func (enc *MMSEncoder) writeByte(b byte) error {
	bytes := []byte{b}
	if n, err := enc.w.Write(bytes); n != 1 {
		return fmt.Errorf("expected to write 1 byte but wrote %d", n)
	} else if err != nil {
		return err
	}
	return nil
}
