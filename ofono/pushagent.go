/*
 * Copyright 2014 Canonical Ltd.
 *
 * Authors:
 * Sergio Schvezov: sergio.schvezov@cannical.com
 *
 * This file is part of nuntium.
 *
 * nuntium is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; version 3.
 *
 * nuntium is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package ofono

import (
	"encoding/hex"
	"fmt"
	"log"
	"sync"

	"launchpad.net/go-dbus/v1"
	"launchpad.net/nuntium/mms"
)

/*
 in = "aya{sv}", out = ""
*/
type OfonoPushNotification struct {
	Data []byte
	Info map[string]*dbus.Variant
}

type PushAgent struct {
	conn           *dbus.Connection
	modem          dbus.ObjectPath
	Push           chan *PushPDU
	messageChannel chan *dbus.Message
	Registered     bool
	m              sync.Mutex
}

func NewPushAgent(modem dbus.ObjectPath) *PushAgent {
	return &PushAgent{modem: modem}
}

func (agent *PushAgent) Register() (err error) {
	agent.m.Lock()
	defer agent.m.Unlock()
	if agent.conn == nil {
		if agent.conn, err = dbus.Connect(dbus.SystemBus); err != nil {
			return err
		}
	}
	if agent.Registered {
		log.Printf("Agent already registered for %s", agent.modem)
		return nil
	}
	agent.Registered = true
	log.Print("Registering agent for ", agent.modem, " on path ", AGENT_TAG, " and name ", agent.conn.UniqueName)
	obj := agent.conn.Object("org.ofono", agent.modem)
	_, err = obj.Call(PUSH_NOTIFICATION_INTERFACE, "RegisterAgent", AGENT_TAG)
	if err != nil {
		return fmt.Errorf("Cannot register agent for %s: %s", agent.modem, err)
	}
	agent.Push = make(chan *PushPDU)
	agent.messageChannel = make(chan *dbus.Message)
	go agent.watchDBusMethodCalls()
	agent.conn.RegisterObjectPath(AGENT_TAG, agent.messageChannel)
	log.Print("Agent Registered for ", agent.modem, " on path ", AGENT_TAG)
	return nil
}

func (agent *PushAgent) Unregister() error {
	agent.m.Lock()
	defer agent.m.Unlock()
	if !agent.Registered {
		log.Print("Agent no registered for %s", agent.modem)
		return nil
	}
	log.Print("Unregistering agent on ", agent.modem)
	obj := agent.conn.Object("org.ofono", agent.modem)
	_, err := obj.Call(PUSH_NOTIFICATION_INTERFACE, "UnregisterAgent", AGENT_TAG)
	if err != nil {
		log.Print("Unregister failed ", err)
		return err
	}
	agent.release()
	agent.modem = dbus.ObjectPath("")
	return nil
}

func (agent *PushAgent) release() {
	agent.Registered = false
	//BUG this seems to not return, but I can't close the channel or panic
	agent.conn.UnregisterObjectPath(AGENT_TAG)
	close(agent.Push)
	agent.Push = nil
	close(agent.messageChannel)
	agent.messageChannel = nil
}

func (agent *PushAgent) watchDBusMethodCalls() {
	var reply *dbus.Message
	for msg := range agent.messageChannel {
		switch {
		case msg.Interface == PUSH_NOTIFICATION_AGENT_INTERFACE && msg.Member == "ReceiveNotification":
			reply = agent.notificationReceived(msg)
		case msg.Interface == PUSH_NOTIFICATION_AGENT_INTERFACE && msg.Member == "Release":
			log.Printf("Push Agent on %s received Release", agent.modem)
			reply = dbus.NewMethodReturnMessage(msg)
			agent.release()
		default:
			log.Print("Received unkown method call on", msg.Interface, msg.Member)
			reply = dbus.NewErrorMessage(msg, "org.freedesktop.DBus.Error.UnknownMethod", "Unknown method")
		}
		if err := agent.conn.Send(reply); err != nil {
			log.Print("Could not send reply: ", err)
		}
	}
}

func (agent *PushAgent) notificationReceived(msg *dbus.Message) (reply *dbus.Message) {
	var push OfonoPushNotification
	if err := msg.Args(&(push.Data), &(push.Info)); err != nil {
		log.Print("Error in received ReceiveNotification() method call ", msg)
		return dbus.NewErrorMessage(msg, "org.freedesktop.DBus.Error", "FormatError")
	} else {
		log.Print("Received ReceiveNotification() method call from ", push.Info["Sender"].Value)
		log.Print("Push data\n", hex.Dump(push.Data))
		dec := NewDecoder(push.Data)
		pdu := new(PushPDU)
		if err := dec.Decode(pdu); err != nil {
			log.Print("Error ", err)
			return dbus.NewErrorMessage(msg, "org.freedesktop.DBus.Error", "DecodeError")
		}
		// TODO later switch on ApplicationId and ContentType to different channels
		if pdu.ApplicationId == mms.PUSH_APPLICATION_ID && pdu.ContentType == mms.VND_WAP_MMS_MESSAGE {
			agent.Push <- pdu
		} else {
			log.Print("Unhandled push pdu", pdu)
		}
		return dbus.NewMethodReturnMessage(msg)
	}
}
