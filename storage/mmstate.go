/*
 * Copyright 2014 Canonical Ltd.
 *
 * Authors:
 * Sergio Schvezov: sergio.schvezov@cannical.com
 *
 * This file is part of telepathy.
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

package storage

import "github.com/ubports/nuntium/mms"

//SendInfo is a map where every key is a destination and the value can be any of:
//
// - "none": no report has been received yet.
// - "expired": recipient did not retrieve the MMS before expiration.
// - "retrieved": MMS successfully retrieved by the recipient.
// - "rejected": recipient rejected the MMS.
// - "deferred": recipient decided to retrieve the MMS at a later time.
// - "indeterminate": cannot determine if the MMS reached its destination.
// - "forwarded": recipient forwarded the MMS without retrieving it first.
// - "unreachable": recipient is not reachable.
type SendInfo map[string]string

//Status represents an MMS' state
//
// Id represents the transaction ID for the MMS if using delivery request reports
//
// State can be:
// - For incoming messages:
//   - NOTIFICATION : m-Notify.Ind PDU not yet downloaded.
//   - DOWNLOADED   : m-Retrieve.Conf PDU downloaded, but not yet communicated to telepathy or acknowledged to MMS provider.
//   - RECEIVED     : m-Retrieve.Conf PDU downloaded and successfully communicated to telepathy, but not acknowledged to MMS provider.
//   - RESPONDED    : m-Retrieve.Conf PDU downloaded and successfully communicated to telepathy and acknowledged to MMS provider.
// - For outgoing messages:
//   - DRAFT : m-Send.Req PDU ready for sending.
//   - SENT  : m-Send.Req PDU successfully sent.
//
// SendState contains the sent state for each delivered message associated to
// a particular MMS
//
// ModemId represents ID of modem to which the message belongs
//
// MNotificationInd holds the received m-Notify.Ind until PDU downloaded (is not nil when State is "notification").
type MMSState struct {
	Id               string
	State            string
	ContentLocation  string
	SendState        SendInfo
	ModemId          string
	MNotificationInd *mms.MNotificationInd
}

func (m MMSState) IsIncoming() bool {
	return m.State == NOTIFICATION || m.State == DOWNLOADED || m.State == RECEIVED || m.State == RESPONDED
}
