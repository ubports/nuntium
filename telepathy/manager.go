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

	"github.com/ubports/nuntium/mms"
	"launchpad.net/go-dbus/v1"
)

type MMSManager struct {
	conn     *dbus.Connection
	msgChan  chan *dbus.Message
	services []*MMSService
}

func NewMMSManager(conn *dbus.Connection) (*MMSManager, error) {
	name := conn.RequestName(MMS_DBUS_NAME, dbus.NameFlagDoNotQueue)
	err := <-name.C
	if err != nil {
		return nil, fmt.Errorf("Could not aquire name %s", MMS_DBUS_NAME)
	}

	log.Printf("Registered %s on bus as %s", conn.UniqueName, name.Name)

	manager := MMSManager{conn: conn, msgChan: make(chan *dbus.Message)}
	go manager.watchDBusMethodCalls()
	conn.RegisterObjectPath(MMS_DBUS_PATH, manager.msgChan)
	return &manager, nil
}

func (manager *MMSManager) watchDBusMethodCalls() {
	var reply *dbus.Message

	for msg := range manager.msgChan {
		switch {
		case msg.Interface == MMS_MANAGER_DBUS_IFACE && msg.Member == "GetServices":
			log.Print("Received GetServices()")
			reply = manager.getServices(msg)
		default:
			log.Println("Received unkown method call on", msg.Interface, msg.Member)
			reply = dbus.NewErrorMessage(msg, "org.freedesktop.DBus.Error.UnknownMethod", "Unknown method")
		}
		if err := manager.conn.Send(reply); err != nil {
			log.Print("Could not send reply: ", err)
		}
	}
}

func (manager *MMSManager) getServices(msg *dbus.Message) *dbus.Message {
	var payloads []Payload
	for i, _ := range manager.services {
		payloads = append(payloads, manager.services[i].payload)
	}
	reply := dbus.NewMethodReturnMessage(msg)
	if err := reply.AppendArgs(payloads); err != nil {
		log.Print("Cannot parse payload data from services")
		return dbus.NewErrorMessage(msg, "Error.InvalidArguments", "Cannot parse services")
	}
	return reply
}

func (manager *MMSManager) serviceAdded(payload *Payload) error {
	log.Print("Service added ", payload.Path)
	signal := dbus.NewSignalMessage(MMS_DBUS_PATH, MMS_MANAGER_DBUS_IFACE, serviceAddedSignal)
	if err := signal.AppendArgs(payload.Path, payload.Properties); err != nil {
		return err
	}
	if err := manager.conn.Send(signal); err != nil {
		return fmt.Errorf("Cannot send ServiceAdded for %s", payload.Path)
	}
	return nil
}

//TODO:version - Change so we don't need to bump major version.
func (manager *MMSManager) AddService(identity string, modemObjPath dbus.ObjectPath, outgoingChannel chan *OutgoingMessage, useDeliveryReports bool, mNotificationIndChan chan<- *mms.MNotificationInd) (*MMSService, error) {
	for i := range manager.services {
		if manager.services[i].isService(identity) {
			return manager.services[i], nil
		}
	}
	service := NewMMSService(manager.conn, modemObjPath, identity, outgoingChannel, useDeliveryReports, mNotificationIndChan)
	if err := manager.serviceAdded(&service.payload); err != nil {
		return &MMSService{}, err
	}
	manager.services = append(manager.services, service)
	return service, nil
}

func (manager *MMSManager) serviceRemoved(payload *Payload) error {
	log.Print("Service removed ", payload.Path)
	signal := dbus.NewSignalMessage(MMS_DBUS_PATH, MMS_MANAGER_DBUS_IFACE, serviceRemovedSignal)
	if err := signal.AppendArgs(payload.Path); err != nil {
		return err
	}
	if err := manager.conn.Send(signal); err != nil {
		return fmt.Errorf("Cannot send ServiceRemoved for %s", payload.Path)
	}
	return nil
}

func (manager *MMSManager) RemoveService(identity string) error {
	for i := range manager.services {
		if manager.services[i].isService(identity) {
			manager.serviceRemoved(&manager.services[i].payload)
			manager.services[i].Close()
			manager.services = append(manager.services[:i], manager.services[i+1:]...)
			log.Print("Service left: ", len(manager.services))
			return nil
		}
	}
	return fmt.Errorf("Cannot find service serving %s", identity)
}
