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

type ContentType struct {
	Name, FileName, CharSet, Start, StartInfo, Domain, Path, Comment string
	Size, Type, CreationDate, ModificationDate, ReadDate, MediaType  uint64
	Secure                                                           bool
}

func (dec *MMSDecoder) readContentType(reflectedPdu *reflect.Value, hdr string) error {
	// Only implementing general form decoding from 8.4.2.24
	var err error
	var length, mediaType uint64
	if length, err = dec.readLength(); err != nil {
		return err
	}
	fmt.Println("length", length)
	end := int(length) + dec.offset
	if mediaType, err = dec.readInteger(nil, ""); err != nil {
		return err
	}
	fmt.Println("Media Type", CONTENT_TYPES[mediaType])
	for dec.offset < len(dec.data) && dec.offset < end {
		fmt.Printf("offset %d, value: %#x\n", dec.offset, dec.data[dec.offset])
		param, _ := dec.readInteger(nil, "")
		//param := dec.data[dec.offset] & 0x7D
		fmt.Printf("offset %d, value: %#x, param %#x\n", dec.offset, dec.data[dec.offset], param)
		switch param {
		case WSP_PARAMETER_TYPE_Q:
			fmt.Println("Unhandled Q")
		case WSP_PARAMETER_TYPE_CHARSET:
			fmt.Println("Unhandled Charset")
		case WSP_PARAMETER_TYPE_LEVEL:
			fmt.Println("Unhandled Level")
		case WSP_PARAMETER_TYPE_TYPE:
			v, _ := dec.readInteger(nil, "")
			fmt.Println("Type", v)
		case WSP_PARAMETER_TYPE_NAME_DEFUNCT:
			v, _ := dec.readString(nil, "")
			fmt.Println("Name(deprecated)", v)
		case WSP_PARAMETER_TYPE_FILENAME_DEFUNCT:
			v, _ := dec.readString(nil, "")
			fmt.Println("FileName(deprecated)", v)
		case WSP_PARAMETER_TYPE_DIFFERENCES:
			fmt.Println("Unhandled Differences")
		case WSP_PARAMETER_TYPE_PADDING:
			dec.readShortInteger(nil, "")
		case WSP_PARAMETER_TYPE_CONTENT_TYPE:
			v, _ := dec.readString(nil, "")
			fmt.Println("Content Type constrained", v)
		case WSP_PARAMETER_TYPE_START_DEFUNCT:
			v, _ := dec.readString(nil, "")
			fmt.Println("Start(deprecated)", v)
		case WSP_PARAMETER_TYPE_START_INFO_DEFUNCT:
			v, _ := dec.readString(nil, "")
			fmt.Println("Start Info(deprecated", v)
		case WSP_PARAMETER_TYPE_COMMENT_DEFUNCT:
			v, _ := dec.readString(nil, "")
			fmt.Println("Comment(deprecated", v)
		case WSP_PARAMETER_TYPE_DOMAIN_DEFUNCT:
			v, _ := dec.readString(nil, "")
			fmt.Println("Domain(deprecated)", v)
		case WSP_PARAMETER_TYPE_MAX_AGE:
			fmt.Println("Unhandled MAX Age")
		case WSP_PARAMETER_TYPE_PATH_DEFUNCT:
			v, _ := dec.readString(nil, "")
			fmt.Println("Path(deprecated)", v)
		case WSP_PARAMETER_TYPE_SECURE:
			fmt.Println("Secure")
		case WSP_PARAMETER_TYPE_SEC:
			v, _ := dec.readShortInteger(nil, "")
			fmt.Println("SEC(deprecated)", v)
		case WSP_PARAMETER_TYPE_MAC:
			fmt.Println("Unhandled MAC")
		case WSP_PARAMETER_TYPE_CREATION_DATE:
			fmt.Println("Unhandled Creation Date")
		case WSP_PARAMETER_TYPE_MODIFICATION_DATE:
			fmt.Println("Unhandled Modification Date")
		case WSP_PARAMETER_TYPE_READ_DATE:
			fmt.Println("Unhandled Read Date")
		case WSP_PARAMETER_TYPE_SIZE:
			v, _ := dec.readInteger(nil, "")
			fmt.Println("Size", v)
		case WSP_PARAMETER_TYPE_NAME:
			v, _ := dec.readString(nil, "")
			fmt.Println("Name", v)
		case WSP_PARAMETER_TYPE_FILENAME:
			v, _ := dec.readString(nil, "")
			fmt.Println("FileName", v)
		case WSP_PARAMETER_TYPE_START:
			v, _ := dec.readString(nil, "")
			fmt.Println("Start", v)
		case WSP_PARAMETER_TYPE_START_INFO:
			v, _ := dec.readString(nil, "")
			fmt.Println("Start Info", v)
		case WSP_PARAMETER_TYPE_COMMENT:
			v, _ := dec.readString(nil, "")
			fmt.Println("Comment", v)
		case WSP_PARAMETER_TYPE_DOMAIN:
			v, _ := dec.readString(nil, "")
			fmt.Println("Domain", v)
		case WSP_PARAMETER_TYPE_PATH:
			v, _ := dec.readString(nil, "")
			fmt.Println("Path", v)
		case WSP_PARAMETER_TYPE_UNTYPED:
			v, _ := dec.readString(nil, "")
			fmt.Println("Untyped", v)
		default:
			fmt.Println("Unhandled")
		}
		if err != nil {
			return err
		}
	}
	return nil
}

