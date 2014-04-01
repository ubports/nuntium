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
)

type ContentType struct {
	Name, Type, FileName, CharSet, Start, StartInfo, Domain, Path, Comment, MediaType string
	Level																			  byte
	Length, Size, CreationDate, ModificationDate, ReadDate                            uint64
	Secure                                                                            bool
	Q                                                                                 float64
}

type DataPart struct {
	ContentType		ContentType
	Data			[]byte
}

func (dec *MMSDecoder) readQ(reflectedPdu *reflect.Value) error {
	v, err := dec.readUintVar(nil, "") 
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

func (dec *MMSDecoder) readLength(reflectedPdu *reflect.Value) (length uint64, err error) {
	switch {
	case dec.data[dec.offset+1] < 30:
		l, err := dec.readShortInteger(nil, "")
		v := uint64(l)
		reflectedPdu.FieldByName("Length").SetUint(v)
		return v, err
	case dec.data[dec.offset+1] == 31:
		dec.offset++

	}
	return 0, fmt.Errorf("Unhandled lenght")
}

func (dec *MMSDecoder) readContentTypeParts(reflectedPdu *reflect.Value) error {
	dec.offset++
	var err error
	var parts uint64
	if parts, err = dec.readUintVar(nil, ""); err != nil {
		return err
	}
	dec.offset++
	fmt.Println("Number of parts", parts)
	for i := uint64(0); i < parts; i++ {
		headerLen, err := dec.readUintVar(nil, "")
		if err != nil {
			return err
		}
		dataLen, err := dec.readUintVar(nil, "")
		if err != nil {
			return err
		}
		fmt.Println("header len:", headerLen, "dataLen:", dataLen)
		var ct ContentType
		ctReflected := reflect.ValueOf(&ct).Elem()
		err = dec.readContentType(&ctReflected, dec.offset + int(headerLen))
		if err != nil {
			return err
		}
		fmt.Println(ct)
	}

	return nil
}

func (dec *MMSDecoder) readContentTypeHeaders(ctMember *reflect.Value) (int, error) {
	var err error
	var length, mediaType uint64
	if length, err = dec.readLength(ctMember); err != nil {
		return 0, err
	}
	fmt.Println("Content Type Length:", length)
	endOffset := int(length) + dec.offset
	if mediaType, err = dec.readInteger(nil, ""); err != nil {
		return 0, err
	}

	//TODO error checking
	ctMember.FieldByName("MediaType").SetString(CONTENT_TYPES[mediaType])
	fmt.Println("Media Type:", CONTENT_TYPES[mediaType])

	return endOffset, nil
}

func (dec *MMSDecoder) readContentType(ctMember *reflect.Value, endOffset int) error {
	// Only implementing general form decoding from 8.4.2.24
	var err error
	for dec.offset < len(dec.data) && dec.offset < endOffset {
		param, _ := dec.readInteger(nil, "")
		fmt.Printf("offset %d, value: %#x, param %#x\n", dec.offset, dec.data[dec.offset], param)
		switch param {
		case WSP_PARAMETER_TYPE_Q:
			err = dec.readQ(ctMember)
		case WSP_PARAMETER_TYPE_CHARSET:
			fmt.Println("Unhandled Charset")
		case WSP_PARAMETER_TYPE_LEVEL:
			_, err = dec.readShortInteger(ctMember, "Level")
		case WSP_PARAMETER_TYPE_TYPE:
			_, err = dec.readInteger(ctMember, "Type")
		case WSP_PARAMETER_TYPE_NAME_DEFUNCT:
			_, err = dec.readString(ctMember, "Name")
			fmt.Println("Name(deprecated)")
		case WSP_PARAMETER_TYPE_FILENAME_DEFUNCT:
			fmt.Println("FileName(deprecated)")
			_, err = dec.readString(ctMember, "FileName")
		case WSP_PARAMETER_TYPE_DIFFERENCES:
			err = errors.New("Unhandled Differences")
		case WSP_PARAMETER_TYPE_PADDING:
			dec.readShortInteger(nil, "")
		case WSP_PARAMETER_TYPE_CONTENT_TYPE:
			_, err = dec.readString(ctMember, "Type")
		case WSP_PARAMETER_TYPE_START_DEFUNCT:
			fmt.Println("Start(deprecated)")
			_, err = dec.readString(ctMember, "Start")
		case WSP_PARAMETER_TYPE_START_INFO_DEFUNCT:
			fmt.Println("Start Info(deprecated")
			_, err = dec.readString(ctMember, "StartInfo")
		case WSP_PARAMETER_TYPE_COMMENT_DEFUNCT:
			fmt.Println("Comment(deprecated")
			_, err = dec.readString(ctMember, "Comment")
		case WSP_PARAMETER_TYPE_DOMAIN_DEFUNCT:
			fmt.Println("Domain(deprecated)")
			_, err = dec.readString(ctMember, "Domain")
		case WSP_PARAMETER_TYPE_MAX_AGE:
			err = errors.New("Unhandled Max Age")
		case WSP_PARAMETER_TYPE_PATH_DEFUNCT:
			fmt.Println("Path(deprecated)")
			_, err = dec.readString(ctMember, "Path")
		case WSP_PARAMETER_TYPE_SECURE:
			fmt.Println("Secure")
		case WSP_PARAMETER_TYPE_SEC:
			v, _ := dec.readShortInteger(nil, "")
			fmt.Println("SEC(deprecated)", v)
		case WSP_PARAMETER_TYPE_MAC:
			err = errors.New("Unhandled MAC")
		case WSP_PARAMETER_TYPE_CREATION_DATE:
		case WSP_PARAMETER_TYPE_MODIFICATION_DATE:
		case WSP_PARAMETER_TYPE_READ_DATE:
			err = errors.New("Unhandled Date parameters")
		case WSP_PARAMETER_TYPE_SIZE:
			_, err = dec.readInteger(ctMember, "Size")
		case WSP_PARAMETER_TYPE_NAME:
			_, err = dec.readString(ctMember, "Name")
		case WSP_PARAMETER_TYPE_FILENAME:
			_, err = dec.readString(ctMember, "FileName")
		case WSP_PARAMETER_TYPE_START:
			_, err = dec.readString(ctMember, "Start")
		case WSP_PARAMETER_TYPE_START_INFO:
			_, err = dec.readString(ctMember, "StartInfo")
		case WSP_PARAMETER_TYPE_COMMENT:
			_, err = dec.readString(ctMember, "Comment")
		case WSP_PARAMETER_TYPE_DOMAIN:
			_, err = dec.readString(ctMember, "Domain")
		case WSP_PARAMETER_TYPE_PATH:
			_, err = dec.readString(ctMember, "Path")
		case WSP_PARAMETER_TYPE_UNTYPED:
			v, _ := dec.readString(nil, "")
			fmt.Println("Untyped", v)
		default:
			err = fmt.Errorf("Unhandled parameter %#x == %d at offset %d", param, param, dec.offset)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
