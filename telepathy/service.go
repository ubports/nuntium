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

import (
	"launchpad.net/go-dbus/v1"
	"log"
)

//ServicePayload is used to build the dbus messages; this is a workaround as v1 of go-dbus
//tries to encode and decode private fields.
type ServicePayload struct {
	Path       dbus.ObjectPath
	Properties map[string]dbus.Variant
}

type MMSService struct {
	Payload    ServicePayload
	Properties map[string]dbus.Variant
	conn       *dbus.Connection
	msgChan    chan *dbus.Message
}

func NewMMSService(conn *dbus.Connection, identity string, useDeliveryReports bool) MMSService {
	properties := make(map[string]dbus.Variant)
	properties[IDENTITY] = dbus.Variant{identity}
	serviceProperties := make(map[string]dbus.Variant)
	serviceProperties[USE_DELIVERY_REPORTS] = dbus.Variant{useDeliveryReports}
	payload := ServicePayload{
		Path:       dbus.ObjectPath(MMS_DBUS_PATH + "/" + identity),
		Properties: properties,
	}
	service := MMSService{
		Payload:    payload,
		Properties: serviceProperties,
		conn:       conn,
		msgChan:    make(chan *dbus.Message),
	}
	go service.watchDBusMethodCalls()
	conn.RegisterObjectPath(payload.Path, service.msgChan)
	return service
}

func (service *MMSService) watchDBusMethodCalls() {
	var reply *dbus.Message

	for msg := range service.msgChan {
		switch {
		case msg.Interface == MMS_SERVICE_DBUS_IFACE && msg.Member == "GetMessages":
			reply = dbus.NewMethodReturnMessage(msg)
			//TODO implement store and forward
			var noMessages []string
			if err := reply.AppendArgs(noMessages); err != nil {
				log.Print("Cannot parse payload data from services")
				reply = dbus.NewErrorMessage(msg, "Error.InvalidArguments", "Cannot parse services")
			}
		case msg.Interface == MMS_SERVICE_DBUS_IFACE && msg.Member == "GetProperties":
			reply = dbus.NewMethodReturnMessage(msg)
			if err := reply.AppendArgs(service.Properties); err != nil {
				log.Print("Cannot parse payload data from services")
				reply = dbus.NewErrorMessage(msg, "Error.InvalidArguments", "Cannot parse services")
			}
		default:
			log.Println("Received unkown method call on", msg.Interface, msg.Member)
			reply = dbus.NewErrorMessage(msg, "org.freedesktop.DBus.Error.UnknownMethod", "Unknown method")
		}
		if err := service.conn.Send(reply); err != nil {
			log.Println("Could not send reply:", err)
		}
	}
}

//MessageAdded emits a MessageAdded with the path to the added message which
//is taken as a parameter
func (service *MMSService) MessageAdded(filePath string) error {
	signal := dbus.NewSignalMessage(service.Payload.Path, MMS_SERVICE_DBUS_IFACE, MESSAGE_ADDED)
	signal.AppendArgs(filePath)
	if err := service.conn.Send(signal); err != nil {
		return err
	}
	return nil
}
