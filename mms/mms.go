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
	"log"
	"os"
	"strings"
	"time"
)

// MMS Field names from OMA-WAP-MMS section 7.3 Table 12
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

// MMS Content Type Assignments OMA-WAP-MMS section 7.3 Table 13
const (
	PUSH_APPLICATION_ID = 4
	VND_WAP_MMS_MESSAGE = "application/vnd.wap.mms-message"
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

// Delivery Report defined in OMA-WAP-MMS section 7.2.6
const (
	DeliveryReportYes byte = 128
	DeliveryReportNo  byte = 129
)

// Expiry tokens defined in OMA-WAP-MMS section 7.2.10
const (
	ExpiryTokenAbsolute byte = 128
	ExpiryTokenRelative byte = 129
)

// From tokens defined in OMA-WAP-MMS section 7.2.11
const (
	TOKEN_ADDRESS_PRESENT = 0x80
	TOKEN_INSERT_ADDRESS  = 0x81
)

// Message classes defined in OMA-WAP-MMS section 7.2.14
const (
	ClassPersonal      byte = 128
	ClassAdvertisement byte = 129
	ClassInformational byte = 130
	ClassAuto          byte = 131
)

// Report Report defined in OMA-WAP-MMS 7.2.20
const (
	ReadReportYes byte = 128
	ReadReportNo  byte = 129
)

// Report Allowed defined in OMA-WAP-MMS section 7.2.26
const (
	ReportAllowedYes byte = 128
	ReportAllowedNo  byte = 129
)

// Response Status defined in OMA-WAP-MMS section 7.2.27
//
// An MMS Client MUST react the same to a value in range 196 to 223 as it
// does to the value 192 (Error-transient-failure).
//
// An MMS Client MUST react the same to a value in range 234 to 255 as it
// does to the value 224 (Error-permanent-failure).
//
// Any other values SHALL NOT be used. They are reserved for future use.
// An MMS Client that receives such a reserved value MUST react the same
// as it does to the value 224 (Error-permanent-failure).
const (
	ResponseStatusOk                            byte = 128
	ResponseStatusErrorUnspecified              byte = 129 // Obsolete
	ResponseStatusErrorServiceDenied            byte = 130 // Obsolete
	ResponseStatusErrorMessageFormatCorrupt     byte = 131 // Obsolete
	ResponseStatusErrorSendingAddressUnresolved byte = 132 // Obsolete
	ResponseStatusErrorMessageNotFound          byte = 133 // Obsolete
	ResponseStatusErrorNetworkProblem           byte = 134 // Obsolete
	ResponseStatusErrorContentNotAccepted       byte = 135 // Obsolete
	ResponseStatusErrorUnsupportedMessage       byte = 136

	ResponseStatusErrorTransientFailure           byte = 192
	ResponseStatusErrorTransientAddressUnresolved byte = 193
	ResponseStatusErrorTransientMessageNotFound   byte = 194
	ResponseStatusErrorTransientNetworkProblem    byte = 195

	ResponseStatusErrorTransientMaxReserved byte = 223

	ResponseStatusErrorPermanentFailure                         byte = 224
	ResponseStatusErrorPermanentServiceDenied                   byte = 225
	ResponseStatusErrorPermanentMessageFormatCorrupt            byte = 226
	ResponseStatusErrorPermanentAddressUnresolved               byte = 227
	ResponseStatusErrorPermanentMessageNotFound                 byte = 228
	ResponseStatusErrorPermanentContentNotAccepted              byte = 229
	ResponseStatusErrorPermanentReplyChargingLimitationsNotMet  byte = 230
	ResponseStatusErrorPermanentReplyChargingRequestNotAccepted byte = 231
	ResponseStatusErrorPermanentReplyChargingForwardingDenied   byte = 232
	ResponseStatusErrorPermanentReplyChargingNotSupported       byte = 233

	ResponseStatusErrorPermamentMaxReserved byte = 255
)

// Status defined in OMA-WAP-MMS section 7.2.23
const (
	STATUS_EXPIRED      = 128
	STATUS_RETRIEVED    = 129
	STATUS_REJECTED     = 130
	STATUS_DEFERRED     = 131
	STATUS_UNRECOGNIZED = 132
)

// MSendReq holds a m-send.req message defined in
// OMA-WAP-MMS-ENC-v1.1 section 6.1.1
type MSendReq struct {
	UUID             string `encode:"no"`
	Type             byte
	TransactionId    string
	Version          byte
	Date             uint64 `encode:"optional"`
	From             string
	To               []string
	Cc               string `encode:"no"`
	Bcc              string `encode:"no"`
	Subject          string `encode:"optional"`
	Class            byte   `encode:"optional"`
	Expiry           uint64 `encode:"optional"`
	DeliveryTime     uint64 `encode:"optional"`
	Priority         byte   `encode:"optional"`
	SenderVisibility byte   `encode:"optional"`
	DeliveryReport   byte   `encode:"optional"`
	ReadReport       byte   `encode:"optional"`
	ContentTypeStart string `encode:"no"`
	ContentTypeType  string `encode:"no"`
	ContentType      string
	Attachments      []*Attachment `encode:"no"`
}

// MSendReq holds a m-send.conf message defined in
// OMA-WAP-MMS-ENC section 6.1.2
type MSendConf struct {
	Type           byte
	TransactionId  string
	Version        byte
	ResponseStatus byte
	ResponseText   string
	MessageId      string
}

// MNotificationInd holds a m-notification.ind message defined in
// OMA-WAP-MMS-ENC section 6.2
type MNotificationInd struct {
	MMSReader
	UUID                                 string
	RedownloadOfUUID                     string // If not empty, it means that the struct was created to redownload a previously failed message download with UUID stored in field.
	Received                             time.Time
	Type, Version, Class, DeliveryReport byte
	ReplyCharging, ReplyChargingDeadline byte
	Priority                             byte
	ReplyChargingId                      string
	TransactionId, ContentLocation       string
	From, Subject                        string
	Expiry                               time.Time
	Size                                 uint64
}

// MNotificationInd holds a m-notifyresp.ind message defined in
// OMA-WAP-MMS-ENC-v1.1 section 6.2
type MNotifyRespInd struct {
	UUID          string `encode:"no"`
	Type          byte
	TransactionId string
	Version       byte
	Status        byte
	ReportAllowed byte `encode:"optional"`
}

// MRetrieveConf holds a m-retrieve.conf message defined in
// OMA-WAP-MMS-ENC-v1.1 section 6.3
type MRetrieveConf struct {
	MMSReader
	UUID                                       string
	Type, Version, Status, Class, Priority     byte
	ReplyCharging, ReplyChargingDeadline       byte
	ReplyChargingId                            string
	ReadReport, RetrieveStatus, DeliveryReport byte
	TransactionId, MessageId, RetrieveText     string
	From, Cc, Subject                          string
	To                                         []string
	ReportAllowed                              byte
	Date                                       uint64
	Content                                    Attachment
	Attachments                                []Attachment
	Data                                       []byte
}

type MMSReader interface{}
type MMSWriter interface{}

// NewMSendReq creates a personal message with a normal priority and no read report
func NewMSendReq(recipients []string, attachments []*Attachment, deliveryReport bool) *MSendReq {
	for i := range recipients {
		recipients[i] += "/TYPE=PLMN"
	}
	uuid := GenUUID()

	orderedAttachments, smilStart, smilType := processAttachments(attachments)

	return &MSendReq{
		Type:          TYPE_SEND_REQ,
		To:            recipients,
		TransactionId: uuid,
		Version:       MMS_MESSAGE_VERSION_1_1,
		UUID:          uuid,
		Date:          getDate(),
		// this will expire the message in 7 days
		Expiry:           uint64(time.Duration(time.Hour * 24 * 7).Seconds()),
		DeliveryReport:   getDeliveryReport(deliveryReport),
		ReadReport:       ReadReportNo,
		Class:            ClassPersonal,
		ContentType:      "application/vnd.wap.multipart.related",
		ContentTypeStart: smilStart,
		ContentTypeType:  smilType,
		Attachments:      orderedAttachments,
	}
}

func NewMSendConf() *MSendConf {
	return &MSendConf{
		Type: TYPE_SEND_CONF,
	}
}

func NewMNotificationInd(received time.Time) *MNotificationInd {
	return &MNotificationInd{Type: TYPE_NOTIFICATION_IND, UUID: GenUUID(), Received: received}
}

func (mNotificationInd *MNotificationInd) IsLocal() bool {
	return strings.HasPrefix(mNotificationInd.ContentLocation, "http://localhost:9191/mms")
}

// Default expire duration is 15 days.
const ExpiryDefaultDuration = 15 * 24 * time.Hour

// Returns the expiry time of the MNotificationInd, which is stored in Expiry field.
// If Expiry field is empty/zero, function returns the time ExpiryDefaultDuration after the time in Received field.
// If both Received and Expiry fields are empty/zero, function returns zero time.
func (mNotificationInd *MNotificationInd) Expire() time.Time {
	if mNotificationInd == nil {
		return time.Time{}
	}
	if mNotificationInd.Expiry.IsZero() {
		if mNotificationInd.Received.IsZero() {
			return time.Time{}
		}
		return mNotificationInd.Received.Add(ExpiryDefaultDuration)
	}
	return mNotificationInd.Expiry
}

// Expiry returns if MNotificationInd is expired at the time of calling this function.
// If both Received and Expiry fields are empty/zero, function returns false.
func (mNotificationInd *MNotificationInd) Expired() bool {
	if mNotificationInd == nil {
		return false
	}
	expire := mNotificationInd.Expire()
	if expire.IsZero() {
		return false
	}
	return time.Now().After(expire)
}

func (mNotificationInd *MNotificationInd) NewMNotifyRespInd(status byte, deliveryReport bool) *MNotifyRespInd {
	return &MNotifyRespInd{
		Type:          TYPE_NOTIFYRESP_IND,
		UUID:          mNotificationInd.UUID,
		TransactionId: mNotificationInd.TransactionId,
		Version:       mNotificationInd.Version,
		Status:        status,
		ReportAllowed: getReportAllowed(deliveryReport),
	}
}

func (mRetrieveConf *MRetrieveConf) NewMNotifyRespInd(deliveryReport bool) *MNotifyRespInd {
	return &MNotifyRespInd{
		Type:          TYPE_NOTIFYRESP_IND,
		UUID:          mRetrieveConf.UUID,
		TransactionId: mRetrieveConf.TransactionId,
		Version:       mRetrieveConf.Version,
		Status:        STATUS_RETRIEVED,
		ReportAllowed: getReportAllowed(deliveryReport),
	}
}

func NewMNotifyRespInd() *MNotifyRespInd {
	return &MNotifyRespInd{Type: TYPE_NOTIFYRESP_IND}
}

func NewMRetrieveConf(uuid string) *MRetrieveConf {
	return &MRetrieveConf{Type: TYPE_RETRIEVE_CONF, UUID: uuid}
}

func GenUUID() string {
	var id string
	random, err := os.Open("/dev/urandom")
	if err != nil {
		id = "1234567890ABCDEF"
	} else {
		defer random.Close()
		b := make([]byte, 16)
		random.Read(b)
		id = fmt.Sprintf("%x", b)
	}
	return id
}

var ErrTransient = errors.New("Error-transient-failure")
var ErrPermanent = errors.New("Error-permament-failure")

func (mSendConf *MSendConf) Status() error {
	s := mSendConf.ResponseStatus
	// these are case by case Response Status and we need to determine each one
	switch s {
	case ResponseStatusOk:
		return nil
	case ResponseStatusErrorUnspecified:
		return ErrTransient
	case ResponseStatusErrorServiceDenied:
		return ErrTransient
	case ResponseStatusErrorMessageFormatCorrupt:
		return ErrPermanent
	case ResponseStatusErrorSendingAddressUnresolved:
		return ErrPermanent
	case ResponseStatusErrorMessageNotFound:
		// this could be ErrTransient or ErrPermanent
		return ErrPermanent
	case ResponseStatusErrorNetworkProblem:
		return ErrTransient
	case ResponseStatusErrorContentNotAccepted:
		return ErrPermanent
	case ResponseStatusErrorUnsupportedMessage:
		return ErrPermanent
	}

	// these are the Response Status we can group
	if s >= ResponseStatusErrorTransientFailure && s <= ResponseStatusErrorTransientMaxReserved {
		return ErrTransient
	} else if s >= ResponseStatusErrorPermanentFailure && s <= ResponseStatusErrorPermamentMaxReserved {
		return ErrPermanent
	}

	// any case not handled is a permanent error
	return ErrPermanent
}

func getReadReport(v bool) (read byte) {
	if v {
		read = ReadReportYes
	} else {
		read = ReadReportNo
	}
	return read
}

func getDeliveryReport(v bool) (delivery byte) {
	if v {
		delivery = DeliveryReportYes
	} else {
		delivery = DeliveryReportNo
	}
	return delivery
}

func getReportAllowed(v bool) (allowed byte) {
	if v {
		allowed = ReportAllowedYes
	} else {
		allowed = ReportAllowedNo
	}
	return allowed
}

func getDate() (date uint64) {
	d := time.Now().Unix()
	if d > 0 {
		date = uint64(d)
	}
	return date
}

func processAttachments(a []*Attachment) (oa []*Attachment, smilStart, smilType string) {
	oa = make([]*Attachment, 0, len(a))
	for i := range a {
		if strings.HasPrefix(a[i].MediaType, "application/smil") {
			oa = append([]*Attachment{a[i]}, oa...)
			var err error
			smilStart, err = getSmilStart(a[i].Data)
			if err != nil {
				log.Println("Cannot set content type start:", err)
			}
			smilType = "application/smil"
		} else {
			oa = append(oa, a[i])
		}
	}
	return oa, smilStart, smilType
}
