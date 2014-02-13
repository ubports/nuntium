/*
 * Copyright 2014 Canonical Ltd.
 *
 * Authors:
 * Sergio Schvezov: sergio.schvezov@cannical.com
 *
 * This file is part of ememesd.
 *
 * ememesd is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; version 3.
 *
 * ememesd is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"errors"
	"fmt"
	"launchpad.net/go-dbus/v1"
	"log"
	"net"
	"reflect"
	"strconv"
)

const (
	AGENT_TAG                         = dbus.ObjectPath("/ememesd")
	PUSH_NOTIFICATION_INTERFACE       = "org.ofono.PushNotification"
	PUSH_NOTIFICATION_AGENT_INTERFACE = "org.ofono.PushNotificationAgent"
)

type PropertiesType map[string]dbus.Variant

type OfonoContext struct {
	ObjectPath dbus.ObjectPath
	Properties PropertiesType
}

/*
 in = "aya{sv}", out = ""
*/
type OfonoPushNotification struct {
	Data []byte
	Info map[string]*dbus.Variant
}

type Modem struct {
	modem           dbus.ObjectPath
	Contexts        []OfonoContext
	agentRegistered bool
	messageChannel  chan *dbus.Message
	mmsChannel      chan *PushPDU
}

type ProxyInfo struct {
	Host string
	Port uint64
}

func (oProp OfonoContext) String() string {
	var s string
	s += fmt.Sprintf("ObjectPath: %s\n", oProp.ObjectPath)
	for k, v := range oProp.Properties {
		s += fmt.Sprint("\t", k, ": ", v.Value, "\n")
	}
	return s
}

func NewModems(conn *dbus.Connection) ([]Modem, error) {
	var modems []Modem
	obj := conn.Object("org.ofono", "/")
	modemsReply, err := getOfonoProps(obj, "org.ofono.Manager", "GetModems")
	if err != nil {
		return modems, err
	}
	for _, modemReply := range modemsReply {
		var modem Modem
		modem.modem = modemReply.ObjectPath
		modems = append(modems, modem)
	}
	return modems, nil
}

func getOfonoProps(obj *dbus.ObjectProxy, iface, method string) (oProps []OfonoContext, err error) {
	reply, err := obj.Call(iface, method)
	if err != nil || reply.Type == dbus.TypeError {
		return oProps, err
	}
	if err := reply.Args(&oProps); err != nil {
		return oProps, err
	}
	return oProps, err
}

func (oContext OfonoContext) GetProxy() (proxyInfo ProxyInfo, err error) {
	proxy := reflect.ValueOf(oContext.Properties["MessageProxy"].Value).String()
	var portString string
	proxyInfo.Host, portString, err = net.SplitHostPort(proxy)
	if err != nil {
		return proxyInfo, err
	}
	proxyInfo.Port, err = strconv.ParseUint(portString, 0, 64)
	if err != nil {
		return proxyInfo, err
	}
	return proxyInfo, nil
}

func (modem *Modem) GetContexts(conn *dbus.Connection, contextType string) error {
	rilObj := conn.Object("org.ofono", modem.modem)
	contexts, err := getOfonoProps(rilObj, "org.ofono.ConnectionManager", "GetContexts")
	if err != nil {
		return err
	}
	for _, context := range contexts {
		for k, v := range context.Properties {
			if reflect.ValueOf(k).Kind() != reflect.String || reflect.ValueOf(v.Value).Kind() != reflect.String {
				continue
			}
			k, v.Value = reflect.ValueOf(k).String(), reflect.ValueOf(v.Value).String()
			if k != "Type" {
				continue
			}
			if v.Value != contextType {
				continue
			}
			modem.Contexts = append(modem.Contexts, context)
		}
	}
	return nil
}

func (modem *Modem) RegisterAgent(conn *dbus.Connection, mmsChannel chan *PushPDU) error {
	if modem.agentRegistered {
		return nil
	}
	obj := conn.Object("org.ofono", modem.modem)
	_, err := obj.Call(PUSH_NOTIFICATION_INTERFACE, "RegisterAgent", AGENT_TAG)
	if err != nil {
		return errors.New(fmt.Sprintf("Cannot register agent for %s: %s", modem.modem, err))
	}
	modem.agentRegistered = true
	modem.mmsChannel = mmsChannel
	modem.messageChannel = make(chan *dbus.Message)
	go modem.watchDBusMethodCalls(conn)
	conn.RegisterObjectPath(AGENT_TAG, modem.messageChannel)
	return nil
}

func (modem *Modem) UnregisterAgent(conn *dbus.Connection) error {
	log.Print("Unregistering agent on ", modem.modem)
	obj := conn.Object("org.ofono", modem.modem)
	_, err := obj.Call(PUSH_NOTIFICATION_INTERFACE, "UnregisterAgent", AGENT_TAG)
	if err != nil {
		log.Print("Unregister failed ", err)
		return err
	}
	conn.UnregisterObjectPath(AGENT_TAG)
	close(modem.messageChannel)
	close(modem.mmsChannel)
	modem.agentRegistered = false
	return nil
}

func (modem *Modem) watchDBusMethodCalls(conn *dbus.Connection) {
	var reply *dbus.Message
	for msg := range modem.messageChannel {
		switch {
		case msg.Interface == PUSH_NOTIFICATION_AGENT_INTERFACE && msg.Member == "ReceiveNotification":
			reply = notificationReceived(conn, msg, modem.mmsChannel)
		case msg.Interface == PUSH_NOTIFICATION_AGENT_INTERFACE && msg.Member == "Release":
			log.Print("Received Release")
			reply = dbus.NewMethodReturnMessage(msg)
		default:
			log.Print("Received unkown method call on", msg.Interface, msg.Member)
			reply = dbus.NewErrorMessage(msg, "org.freedesktop.DBus.Error.UnknownMethod", "Unknown method")
		}
		if err := conn.Send(reply); err != nil {
			log.Print("Could not send reply: ", err)
		}
	}
}

func notificationReceived(conn *dbus.Connection, msg *dbus.Message, pushChannel chan *PushPDU) (reply *dbus.Message) {
	var push OfonoPushNotification
	if err := msg.Args(&(push.Data), &(push.Info)); err != nil {
		log.Print("Error in received ReceiveNotification() method call ", msg)
		return dbus.NewErrorMessage(msg, "org.freedesktop.DBus.Error", "FormatError")
	} else {
		log.Print("Received ReceiveNotification() method call from ", push.Info["Sender"].Value)
		log.Printf("Push data %x", push.Data)
		dec := NewDecoder(push.Data)
		pdu := new(PushPDU)
		if err := dec.Decode(pdu); err != nil {
			log.Print("Error ", err)
			return dbus.NewErrorMessage(msg, "org.freedesktop.DBus.Error", "DecodeError")
		}
		// TODO later switch on ApplicationId and ContentType to different channels
		if pdu.ApplicationId == 0x04 && pdu.ContentType == "application/vnd.wap.mms-message" {
			pushChannel <- pdu
		} else {
			log.Print("Unhandled push pdu", pdu)
		}
		return dbus.NewMethodReturnMessage(msg)
	}
}
