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
	"fmt"
	"log"
	"sort"

	"launchpad.net/go-dbus/v1"
)

var validStatus sort.StringSlice

func init() {
	validStatus = sort.StringSlice{SENT, PERMANENT_ERROR, TRANSIENT_ERROR}
	sort.Strings(validStatus)
}

type MessageInterface struct {
	conn       *dbus.Connection
	objectPath dbus.ObjectPath
	msgChan    chan *dbus.Message
	deleteChan chan dbus.ObjectPath
	status     string
}

func NewMessageInterface(conn *dbus.Connection, objectPath dbus.ObjectPath, deleteChan chan dbus.ObjectPath) *MessageInterface {
	msgInterface := MessageInterface{
		conn:       conn,
		objectPath: objectPath,
		deleteChan: deleteChan,
		msgChan:    make(chan *dbus.Message),
		status:     "draft",
	}
	go msgInterface.watchDBusMethodCalls()
	conn.RegisterObjectPath(msgInterface.objectPath, msgInterface.msgChan)
	return &msgInterface
}

func (msgInterface *MessageInterface) Close() {
	close(msgInterface.msgChan)
	msgInterface.msgChan = nil
	msgInterface.conn.UnregisterObjectPath(msgInterface.objectPath)
}

func (msgInterface *MessageInterface) watchDBusMethodCalls() {
	var reply *dbus.Message

	for msg := range msgInterface.msgChan {
		if msg.Interface != MMS_MESSAGE_DBUS_IFACE {
			log.Println("Received unkown method call on", msg.Interface, msg.Member)
			reply = dbus.NewErrorMessage(msg, "org.freedesktop.DBus.Error.UnknownMethod", "Unknown method")
			continue
		}
		switch msg.Member {
		case "Delete":
			reply = dbus.NewMethodReturnMessage(msg)
			//TODO implement store and forward
			if err := msgInterface.conn.Send(reply); err != nil {
				log.Println("Could not send reply:", err)
			}
			msgInterface.deleteChan <- msgInterface.objectPath
		default:
			log.Println("Received unkown method call on", msg.Interface, msg.Member)
			reply = dbus.NewErrorMessage(msg, "org.freedesktop.DBus.Error.UnknownMethod", "Unknown method")
			if err := msgInterface.conn.Send(reply); err != nil {
				log.Println("Could not send reply:", err)
			}
		}
	}
}

func (msgInterface *MessageInterface) StatusChanged(status string) error {
	i := validStatus.Search(status)
	if i < validStatus.Len() && validStatus[i] == status {
		msgInterface.status = status
		signal := dbus.NewSignalMessage(msgInterface.objectPath, MMS_MESSAGE_DBUS_IFACE, PROPERTY_CHANGED)
		if err := signal.AppendArgs(STATUS, dbus.Variant{status}); err != nil {
			return err
		}
		if err := msgInterface.conn.Send(signal); err != nil {
			return err
		}
		log.Print("Status changed for ", msgInterface.objectPath, " to ", status)

	}
	return fmt.Errorf("status %s is not a valid status", status)
}

func (msgInterface *MessageInterface) GetPayload() *Payload {
	properties := make(map[string]dbus.Variant)
	properties["Status"] = dbus.Variant{msgInterface.status}
	return &Payload{
		Path:       msgInterface.objectPath,
		Properties: properties,
	}
}
