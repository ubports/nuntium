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
	"reflect"
)

type MMSEncoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *MMSEncoder {
	return &MMSEncoder{w}
}

func (enc *MMSEncoder) WriteString(s string) error {
	bytes := []byte(s)
	bytes = append(bytes, 0)
	_, err := enc.w.Write(bytes)
	return err
}

func (enc *MMSEncoder) WriteByte(b byte) error {
	bytes := []byte{b}
	if n, err := enc.w.Write(bytes); n != 1 {
		return fmt.Errorf("expected to write 1 byte but wrote %d", n)
	} else if err != nil {
		return err
	}
	return nil
}

func (enc *MMSEncoder) setParam(param byte) error {
	return enc.WriteByte(param | 0x80)
}

func (enc *MMSEncoder) Encode(pdu MMSWriter) error {
	rPdu := reflect.ValueOf(pdu).Elem()

	//Message Type
	msgType := byte(rPdu.FieldByName("Type").Uint())
	if err := enc.setParam(X_MMS_MESSAGE_TYPE); err != nil {
		return err
	}
	if err := enc.WriteByte(msgType); err != nil {
		return err
	}

	//TransactionId
	transactionId := rPdu.FieldByName("TransactionId").String()
	if err := enc.setParam(X_MMS_TRANSACTION_ID); err != nil {
		return err
	}
	if err := enc.WriteString(transactionId); err != nil {
		return err
	}

	//Version
	version := byte(rPdu.FieldByName("Version").Uint())
	if err := enc.setParam(X_MMS_MMS_VERSION); err != nil {
		return err
	}
	if err := enc.WriteByte(version); err != nil {
		return err
	}

	//Status
	if f := rPdu.FieldByName("Status"); f.IsValid() {
		status := byte(f.Uint())
		if err := enc.setParam(X_MMS_STATUS); err != nil {
			return err
		}
		if err := enc.WriteByte(status); err != nil {
			return err
		}
	}
	//ReportAllowed
	if f := rPdu.FieldByName("ReportAllowed"); f.IsValid() {
		reportAllowed := f.Bool()
		enc.setReportAllowed(reportAllowed)
	}
	return nil
}

func (enc *MMSEncoder) setReportAllowed(reportAllowed bool) error {
	if err := enc.setParam(X_MMS_REPORT_ALLOWED); err != nil {
		return err
	}
	var b byte
	if reportAllowed {
		b = REPORT_ALLOWED_YES
	} else {
		b = REPORT_ALLOWED_NO
	}
	if err := enc.WriteByte(b); err != nil {
		return err
	}
	return nil
}
