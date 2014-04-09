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
	"errors"
	"fmt"
	"launchpad.net/go-dbus/v1"
	"log"
	"net"
	"reflect"
	"strconv"
	"strings"
)

const (
	AGENT_TAG                         = dbus.ObjectPath("/nuntium")
	PUSH_NOTIFICATION_INTERFACE       = "org.ofono.PushNotification"
	PUSH_NOTIFICATION_AGENT_INTERFACE = "org.ofono.PushNotificationAgent"
	CONNECTION_MANAGER_INTERFACE      = "org.ofono.ConnectionManager"
	CONNECTION_CONTEXT_INTERFACE      = "org.ofono.ConnectionContext"
	SIM_MANAGER_INTERFACE             = "org.ofono.SimManager"
	OFONO_SENDER                      = "org.ofono"
	MODEM_INTERFACE                   = "org.ofono.Modem"
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
	agentRegistered bool
	messageChannel  chan *dbus.Message
	PushChannel     chan *PushEvent
	identity        string
	IdentityAdded   chan string
	IdentityRemoved chan string
	ReadySignal     chan bool
	ready           bool
}

type PushEvent struct {
	PDU   *PushPDU
	Modem *Modem
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

func NewModems(conn *dbus.Connection) ([]*Modem, error) {
	var modems []*Modem
	obj := conn.Object("org.ofono", "/")
	modemsReply, err := getOfonoProps(obj, "org.ofono.Manager", "GetModems")
	if err != nil {
		return modems, err
	}
	for _, modemReply := range modemsReply {
		modem := new(Modem)
		modem.modem = modemReply.ObjectPath
		modem.ReadySignal = make(chan bool)
		modem.IdentityAdded = make(chan string)
		modem.IdentityRemoved = make(chan string)
		if err := modem.watchPushInterface(conn, modemReply.Properties); err != nil {
			return modems, err
		}
		modems = append(modems, modem)
	}
	return modems, nil
}

func (modem *Modem) watchPushInterface(conn *dbus.Connection, prop PropertiesType) error {
	if availInterfaces, ok := prop["Interfaces"]; ok {
		availableInterfaces := reflect.ValueOf(availInterfaces.Value)
		for i := 0; i < availableInterfaces.Len(); i++ {
			interfaceName := reflect.ValueOf(availableInterfaces.Index(i).Interface().(string)).String()
			if interfaceName == PUSH_NOTIFICATION_INTERFACE {
				modem.ready = true
				modem.ReadySignal <- modem.ready
				break
			}
		}
	}
	propModemSignal, err := connectToPropertySignal(conn, modem.modem, MODEM_INTERFACE)
	if err != nil {
		return err
	}
	go func() {
		log.Print("Watching push interface to see if we can Register an Agent")
		var propName string
		var propValue dbus.Variant
		var pushInterfaceAvailable, online bool
		//TODO once we start tracking modems, these watches need to become members
		for msg := range propModemSignal.C {
			if err := msg.Args(&propName, &propValue); err != nil {
				log.Print(err)
				continue
			}
			if modem.ready {
				continue
			}
			switch propName {
			case "Interfaces":
				availableInterfaces := reflect.ValueOf(propValue.Value)
				pushInterfaceAvailable = false
				for i := 0; i < availableInterfaces.Len(); i++ {
					interfaceName := reflect.ValueOf(availableInterfaces.Index(i).Interface().(string)).String()
					if interfaceName == PUSH_NOTIFICATION_INTERFACE {
						pushInterfaceAvailable = true
						break
					}
				}
			case "Online":
				online = reflect.ValueOf(propValue.Value).Bool()
			default:
				continue
			}
			if pushInterfaceAvailable && online {
				modem.ready = true
				modem.ReadySignal <- modem.ready
			}
		}
	}()
	return nil
}

func connectToPropertySignal(conn *dbus.Connection, path dbus.ObjectPath, inter string) (*dbus.SignalWatch, error) {
	w, err := conn.WatchSignal(&dbus.MatchRule{
		Type:      dbus.TypeSignal,
		Sender:    OFONO_SENDER,
		Interface: inter,
		Member:    "PropertyChanged",
		Path:      path})
	return w, err
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

//ActivateMMSContext activates a context if necessary and returns the context
//to operate with MMS.
//
//If the context is already active it's a nop.
//Returns either the type=internet context or the type=mms, if none is found
//an error is returned.
func (modem *Modem) ActivateMMSContext(conn *dbus.Connection) (OfonoContext, error) {
	context, err := modem.GetMMSContext(conn)
	if err != nil {
		return OfonoContext{}, err
	}
	if context.isActive() {
		return context, nil
	}
	obj := conn.Object("org.ofono", context.ObjectPath)
	_, err = obj.Call(CONNECTION_CONTEXT_INTERFACE, "SetProperty", "Active", dbus.Variant{true})
	if err != nil {
		return OfonoContext{}, fmt.Errorf("Cannot Activate interface on %s: %s", context.ObjectPath, err)
	}
	return context, nil
}

func (oContext OfonoContext) isActive() bool {
	return reflect.ValueOf(oContext.Properties["Active"].Value).Bool()
}

func (oContext OfonoContext) GetProxy() (proxyInfo ProxyInfo, err error) {
	proxy := reflect.ValueOf(oContext.Properties["MessageProxy"].Value).String()
	if strings.HasPrefix(proxy, "http://") {
		proxy = proxy[len("http://"):]
	}
	var portString string
	proxyInfo.Host, portString, err = net.SplitHostPort(proxy)
	if err != nil {
		proxyInfo.Host = proxy
		proxyInfo.Port = 80
		fmt.Println("Setting proxy to:", proxyInfo)
		return proxyInfo, nil
	}
	proxyInfo.Port, err = strconv.ParseUint(portString, 0, 64)
	if err != nil {
		return proxyInfo, err
	}
	return proxyInfo, nil
}

//GetMMSContexts returns the contexts that are MMS capable; by convention it has
//been defined that for it to be MMS capable it either has to define a MessageProxy
//and a MessageCenter within the context.
//
//The following rules take place:
//- check current type=internet context for MessageProxy & MessageCenter;
//  if they exist and aren't empty AND the context is active, select it as the
//  context to use for MMS.
//- otherwise search for type=mms, if found, use it and activate
//
//Returns either the type=internet context or the type=mms, if none is found
//an error is returned.
func (modem *Modem) GetMMSContext(conn *dbus.Connection) (OfonoContext, error) {
	rilObj := conn.Object("org.ofono", modem.modem)
	contexts, err := getOfonoProps(rilObj, CONNECTION_MANAGER_INTERFACE, "GetContexts")
	if err != nil {
		return OfonoContext{}, err
	}
	for _, context := range contexts {
		var contextType, msgCenter, msgProxy string
		var active bool
		for k, v := range context.Properties {
			if reflect.ValueOf(k).Kind() != reflect.String {
				continue
			}
			k = reflect.ValueOf(k).String()
			switch k {
			case "Type":
				contextType = reflect.ValueOf(v.Value).String()
			case "MessageCenter":
				msgCenter = reflect.ValueOf(v.Value).String()
			case "MessageProxy":
				msgProxy = reflect.ValueOf(v.Value).String()
			case "Active":
				active = reflect.ValueOf(v.Value).Bool()
			}
		}
		log.Println("Context type:", contextType,
			"MessageCenter:", msgCenter,
			"MessageProxy:", msgProxy,
			"Active:", active)
		if contextType == "internet" && active && msgProxy != "" && msgCenter != "" {
			return context, nil
		} else if contextType == "mms" {
			return context, nil
		}
	}
	return OfonoContext{}, errors.New("No mms contexts found")
}

func (modem *Modem) GetIdentity(conn *dbus.Connection) error {
	//wait for the push interface to be available
	defaultError := fmt.Errorf("Cannot retrieve SubscriberIdentity for %s", modem.modem)
	rilObj := conn.Object("org.ofono", modem.modem)
	reply, err := rilObj.Call(SIM_MANAGER_INTERFACE, "GetProperties")
	if err == nil {
		var properties PropertiesType
		if err := reply.Args(&properties); err != nil {
			return defaultError
		}
		if identityVariant, ok := properties["SubscriberIdentity"]; ok {
			modem.identity = reflect.ValueOf(identityVariant.Value).String()
		}
		if modem.identity != "" {
			log.Print("Updating id ", modem.identity)
			modem.IdentityAdded <- modem.identity
		}
	}
	propSimSignal, err := connectToPropertySignal(conn, modem.modem, SIM_MANAGER_INTERFACE)
	if err != nil {
		return err
	}
	go func() {
		var propName string
		var propValue dbus.Variant
		for msg := range propSimSignal.C {
			if err := msg.Args(&propName, &propValue); err != nil {
				log.Print(err)
				continue
			}
			if propName != "SubscriberIdentity" {
				continue
			}
			identity := reflect.ValueOf(propValue.Value).String()
			if identity == "" && modem.identity != "" {
				fmt.Println("Identity before remove", modem.identity)

				modem.IdentityRemoved <- identity
				modem.identity = identity
			}
			fmt.Println("Identity before add", modem.identity)
			if identity != "" && modem.identity == "" {
				//wait for the push interface to be available
				modem.identity = identity
				modem.IdentityAdded <- identity
			}
		}
	}()
	return nil
}

func (modem *Modem) RegisterAgent(conn *dbus.Connection) error {
	if modem.agentRegistered {
		log.Print("Agent already registered")
		return nil
	}
	//it seems that ofono is still doing stuff after advertising the interface as available
	//we will need to sleep a bit
	log.Print("Registering Agent for", modem.modem, " on path ", AGENT_TAG)
	obj := conn.Object("org.ofono", modem.modem)
	_, err := obj.Call(PUSH_NOTIFICATION_INTERFACE, "RegisterAgent", AGENT_TAG)
	if err != nil {
		return fmt.Errorf("Cannot register agent for %s: %s", modem.modem, err)
	}
	modem.agentRegistered = true
	modem.PushChannel = make(chan *PushEvent)
	modem.messageChannel = make(chan *dbus.Message)
	go modem.watchDBusMethodCalls(conn)
	conn.RegisterObjectPath(AGENT_TAG, modem.messageChannel)
	log.Print("Agent Registered for ", modem.modem, " on path ", AGENT_TAG)
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
	modem.Reset(conn)
	return nil
}

func (modem *Modem) Reset(conn *dbus.Connection) {
	log.Print("Resetting modem")
	conn.UnregisterObjectPath(AGENT_TAG)
	modem.agentRegistered = false
	close(modem.PushChannel)
	close(modem.messageChannel)
	modem.IdentityRemoved <- modem.identity
	modem.identity = ""
	modem.ready = false
}

func (modem *Modem) watchDBusMethodCalls(conn *dbus.Connection) {
	var reply *dbus.Message
	for msg := range modem.messageChannel {
		switch {
		case msg.Interface == PUSH_NOTIFICATION_AGENT_INTERFACE && msg.Member == "ReceiveNotification":
			reply = modem.notificationReceived(msg)
		case msg.Interface == PUSH_NOTIFICATION_AGENT_INTERFACE && msg.Member == "Release":
			log.Print("Received Release")
			reply = dbus.NewMethodReturnMessage(msg)
			fmt.Println("Identity before reset", modem.identity)
			modem.Reset(conn)
		default:
			log.Print("Received unkown method call on", msg.Interface, msg.Member)
			reply = dbus.NewErrorMessage(msg, "org.freedesktop.DBus.Error.UnknownMethod", "Unknown method")
		}
		if err := conn.Send(reply); err != nil {
			log.Print("Could not send reply: ", err)
		}
	}
}

func (modem *Modem) notificationReceived(msg *dbus.Message) (reply *dbus.Message) {
	var push OfonoPushNotification
	if err := msg.Args(&(push.Data), &(push.Info)); err != nil {
		log.Print("Error in received ReceiveNotification() method call ", msg)
		return dbus.NewErrorMessage(msg, "org.freedesktop.DBus.Error", "FormatError")
	} else {
		log.Print("Received ReceiveNotification() method call from ", push.Info["Sender"].Value)
		log.Printf("Push data %#x", push.Data)
		dec := NewDecoder(push.Data)
		pdu := new(PushPDU)
		if err := dec.Decode(pdu); err != nil {
			log.Print("Error ", err)
			return dbus.NewErrorMessage(msg, "org.freedesktop.DBus.Error", "DecodeError")
		}
		// TODO later switch on ApplicationId and ContentType to different channels
		if pdu.ApplicationId == 0x04 && pdu.ContentType == "application/vnd.wap.mms-message" {
			modem.PushChannel <- &PushEvent{PDU: pdu, Modem: modem}
		} else {
			log.Print("Unhandled push pdu", pdu)
		}
		return dbus.NewMethodReturnMessage(msg)
	}
}
