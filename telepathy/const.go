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

package telepathy

const (
	MMS_DBUS_NAME          = "org.ofono.mms"
	MMS_DBUS_PATH          = "/org/ofono/mms"
	MMS_MESSAGE_DBUS_IFACE = "org.ofono.mms.Message"
	MMS_SERVICE_DBUS_IFACE = "org.ofono.mms.Service"
	MMS_MANAGER_DBUS_IFACE = "org.ofono.mms.Manager"
)

const (
	identityProperty           string = "Identity"
	useDeliveryReportsProperty string = "UseDeliveryReports"
	modemObjectPathProperty    string = "ModemObjectPath"
	messageAddedSignal         string = "MessageAdded"
	messageRemovedSignal       string = "MessageRemoved"
	serviceAddedSignal         string = "ServiceAdded"
	serviceRemovedSignal       string = "ServiceRemoved"
	preferredContextProperty   string = "PreferredContext"
	propertyChangedSignal      string = "PropertyChanged"
	statusProperty             string = "Status"
)

const (
	PERMANENT_ERROR = "PermanentError"
	SENT            = "Sent"
	TRANSIENT_ERROR = "TransientError"
)

const (
	PLMN = "/TYPE=PLMN"
)
