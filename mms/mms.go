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
	Type, Version, Class, DeliveryReport byte
	ReplyCharging, ReplyChargingDeadline byte
	ReplyChargingId                      string
	TransactionId, ContentLocation       string
	From, Subject                        string
	Expiry, Size                         uint64
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
	Type, Version, Status, Class, Priority     byte
	ReplyCharging, ReplyChargingDeadline       byte
	ReplyChargingId                            string
	ReadReport, RetrieveStatus, DeliveryReport byte
	TransactionId, MessageId, RetrieveText     string
	From, To, Cc, Subject                      string
	ReportAllowed                              bool
	Date                                       uint64
	ContentType                                ContentType
	DataParts                                  []ContentType
	Data                                       []byte
}

type MMSReader interface{}

func NewMNotificationInd() *MNotificationInd {
	return &MNotificationInd{Type: TYPE_NOTIFICATION_IND}
}

func NewMRetrieveConf(filePath string) *MRetrieveConf {
	return &MRetrieveConf{Type: TYPE_RETRIEVE_CONF}
}
