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
	"io/ioutil"
	"log"
	"reflect"
	"strings"
)

type Attachment struct {
	MediaType        string
	Type             string `encode:"no"`
	Name             string
	FileName         string `encode:"no"`
	Charset          string `encode:"no"`
	Start            string
	StartInfo        string `encode:"no"`
	Domain           string `encode:"no"`
	Path             string `encode:"no"`
	Comment          string `encode:"no"`
	ContentLocation  string
	ContentId        string
	Level            byte    `encode:"no"`
	Length           uint64  `encode:"no"`
	Size             uint64  `encode:"no"`
	CreationDate     uint64  `encode:"no"`
	ModificationDate uint64  `encode:"no"`
	ReadDate         uint64  `encode:"no"`
	Offset           int     `encode:"no"`
	Secure           bool    `encode:"no"`
	Q                float64 `encode:"no"`
	Data             []byte  `encode:"no"`
}

func NewAttachment(id, contentType, filePath string) (*Attachment, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot create new ContentType for %s of content type %s on %s: %s", id, contentType, filePath, err)
	}

	ct := &Attachment{
		ContentId:       id,
		ContentLocation: id,
		Data:            data,
	}

	if parts := strings.Split(contentType, ";"); len(parts) > 1 {
		ct.MediaType = parts[0]
		ct.Charset = parts[1]
	} else {
		ct.MediaType = contentType
	}

	if contentType == "application/smil" {
		start, err := getSmilStart(data)
		if err != nil {
			return nil, err
		}
		ct.Start = start
	}
	return ct, nil
}

func getSmilStart(smilData []byte) (string, error) {
	smilStart := string(smilData)

	i := strings.Index(smilStart, ">")
	if i == -1 {
		return "", errors.New("cannot find the SMIL Start tag")
	} else if i+1 > len(smilData) {
		return "", errors.New("buffer overrun while searching for the SMIL Start tag")
	}
	return smilStart[:i+1], nil
}

//GetSmil returns the text corresponding to the ContentType that holds the SMIL
func (pdu *MRetrieveConf) GetSmil() (string, error) {
	for i := range pdu.Attachments {
		if strings.HasPrefix(pdu.Attachments[i].MediaType, "application/smil") {
			return string(pdu.Attachments[i].Data), nil
		}
	}
	return "", errors.New("cannot find SMIL data part")
}

//GetDataParts returns the non SMIL ContentType data parts
func (pdu *MRetrieveConf) GetDataParts() []Attachment {
	var dataParts []Attachment
	for i := range pdu.Attachments {
		if pdu.Attachments[i].MediaType == "application/smil" {
			continue
		}
		dataParts = append(dataParts, pdu.Attachments[i])
	}
	return dataParts
}

func (dec *MMSDecoder) ReadContentTypeParts(reflectedPdu *reflect.Value) error {
	var err error
	var parts uint64
	if parts, err = dec.ReadUintVar(nil, ""); err != nil {
		return err
	}
	var dataParts []Attachment
	dec.log = dec.log + fmt.Sprintf("Number of parts: %d\n", parts)
	for i := uint64(0); i < parts; i++ {
		headerLen, err := dec.ReadUintVar(nil, "")
		if err != nil {
			return err
		}
		dataLen, err := dec.ReadUintVar(nil, "")
		if err != nil {
			return err
		}
		headerEnd := dec.Offset + int(headerLen)
		dec.log = dec.log + fmt.Sprintf("Attachament len(header): %d - len(data) %d\n", headerLen, dataLen)
		var ct Attachment
		ct.Offset = headerEnd + 1
		ctReflected := reflect.ValueOf(&ct).Elem()
		if err := dec.ReadContentType(&ctReflected); err == nil {
			if err := dec.ReadMMSHeaders(&ctReflected, headerEnd); err != nil {
				return err
			}
		} else if err != nil && err.Error() != "WAP message" { //TODO create error type
			return err
		}
		dec.Offset = headerEnd + 1
		if _, err := dec.ReadBoundedBytes(&ctReflected, "Data", dec.Offset+int(dataLen)); err != nil {
			return err
		}
		if ct.MediaType == "application/smil" || strings.HasPrefix(ct.MediaType, "text/plain") || ct.MediaType == "" {
			dec.log = dec.log + fmt.Sprintf("%s\n", ct.Data)
		}
		if ct.Charset != "" {
			ct.MediaType = ct.MediaType + ";charset=" + ct.Charset
		}
		dataParts = append(dataParts, ct)
	}
	dataPartsR := reflect.ValueOf(dataParts)
	reflectedPdu.FieldByName("Attachments").Set(dataPartsR)

	return nil
}

func (dec *MMSDecoder) ReadMMSHeaders(ctMember *reflect.Value, headerEnd int) error {
	for dec.Offset < headerEnd {
		var err error
		param, _ := dec.ReadInteger(nil, "")
		switch param {
		case MMS_PART_CONTENT_LOCATION:
			_, err = dec.ReadString(ctMember, "ContentLocation")
		case MMS_PART_CONTENT_ID:
			_, err = dec.ReadString(ctMember, "ContentId")
		default:
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (dec *MMSDecoder) ReadContentType(ctMember *reflect.Value) error {
	if dec.Offset+1 >= len(dec.Data) {
		return fmt.Errorf("message ended prematurely, offset: %d and payload length is %d", dec.Offset, len(dec.Data))
	}
	// These call the same function
	if next := dec.Data[dec.Offset+1]; next&SHORT_FILTER != 0 {
		return dec.ReadMediaType(ctMember, "MediaType")
	} else if next >= TEXT_MIN && next <= TEXT_MAX {
		return dec.ReadMediaType(ctMember, "MediaType")
	}

	var err error
	var length uint64
	if length, err = dec.ReadLength(ctMember); err != nil {
		return err
	}
	dec.log = dec.log + fmt.Sprintf("Content Type Length: %d\n", length)
	endOffset := int(length) + dec.Offset

	if err := dec.ReadMediaType(ctMember, "MediaType"); err != nil {
		return err
	}

	for dec.Offset < len(dec.Data) && dec.Offset < endOffset {
		param, _ := dec.ReadInteger(nil, "")
		switch param {
		case WSP_PARAMETER_TYPE_Q:
			err = dec.ReadQ(ctMember)
		case WSP_PARAMETER_TYPE_CHARSET:
			_, err = dec.ReadCharset(ctMember, "Charset")
		case WSP_PARAMETER_TYPE_LEVEL:
			_, err = dec.ReadShortInteger(ctMember, "Level")
		case WSP_PARAMETER_TYPE_TYPE:
			_, err = dec.ReadInteger(ctMember, "Type")
		case WSP_PARAMETER_TYPE_NAME_DEFUNCT:
			log.Println("Using deprecated Name header")
			_, err = dec.ReadString(ctMember, "Name")
		case WSP_PARAMETER_TYPE_FILENAME_DEFUNCT:
			log.Println("Using deprecated FileName header")
			_, err = dec.ReadString(ctMember, "FileName")
		case WSP_PARAMETER_TYPE_DIFFERENCES:
			err = errors.New("Unhandled Differences")
		case WSP_PARAMETER_TYPE_PADDING:
			dec.ReadShortInteger(nil, "")
		case WSP_PARAMETER_TYPE_CONTENT_TYPE:
			_, err = dec.ReadString(ctMember, "Type")
		case WSP_PARAMETER_TYPE_START_DEFUNCT:
			log.Println("Using deprecated Start header")
			_, err = dec.ReadString(ctMember, "Start")
		case WSP_PARAMETER_TYPE_START_INFO_DEFUNCT:
			log.Println("Using deprecated StartInfo header")
			_, err = dec.ReadString(ctMember, "StartInfo")
		case WSP_PARAMETER_TYPE_COMMENT_DEFUNCT:
			log.Println("Using deprecated Comment header")
			_, err = dec.ReadString(ctMember, "Comment")
		case WSP_PARAMETER_TYPE_DOMAIN_DEFUNCT:
			log.Println("Using deprecated Domain header")
			_, err = dec.ReadString(ctMember, "Domain")
		case WSP_PARAMETER_TYPE_MAX_AGE:
			err = errors.New("Unhandled Max Age")
		case WSP_PARAMETER_TYPE_PATH_DEFUNCT:
			log.Println("Using deprecated Path header")
			_, err = dec.ReadString(ctMember, "Path")
		case WSP_PARAMETER_TYPE_SECURE:
			log.Println("Unhandled Secure header detected")
		case WSP_PARAMETER_TYPE_SEC:
			v, _ := dec.ReadShortInteger(nil, "")
			log.Println("Using deprecated and unhandled Sec header with value", v)
		case WSP_PARAMETER_TYPE_MAC:
			err = errors.New("Unhandled MAC")
		case WSP_PARAMETER_TYPE_CREATION_DATE:
		case WSP_PARAMETER_TYPE_MODIFICATION_DATE:
		case WSP_PARAMETER_TYPE_READ_DATE:
			err = errors.New("Unhandled Date parameters")
		case WSP_PARAMETER_TYPE_SIZE:
			_, err = dec.ReadInteger(ctMember, "Size")
		case WSP_PARAMETER_TYPE_NAME:
			_, err = dec.ReadString(ctMember, "Name")
		case WSP_PARAMETER_TYPE_FILENAME:
			_, err = dec.ReadString(ctMember, "FileName")
		case WSP_PARAMETER_TYPE_START:
			_, err = dec.ReadString(ctMember, "Start")
		case WSP_PARAMETER_TYPE_START_INFO:
			_, err = dec.ReadString(ctMember, "StartInfo")
		case WSP_PARAMETER_TYPE_COMMENT:
			_, err = dec.ReadString(ctMember, "Comment")
		case WSP_PARAMETER_TYPE_DOMAIN:
			_, err = dec.ReadString(ctMember, "Domain")
		case WSP_PARAMETER_TYPE_PATH:
			_, err = dec.ReadString(ctMember, "Path")
		case WSP_PARAMETER_TYPE_UNTYPED:
			v, _ := dec.ReadString(nil, "")
			log.Println("Unhandled Secure header detected with value", v)
		default:
			err = fmt.Errorf("Unhandled parameter %#x == %d at offset %d", param, param, dec.Offset)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
