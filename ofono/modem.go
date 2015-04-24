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
	"log"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"

	"launchpad.net/go-dbus/v1"
)

const (
	contextTypeInternet = "internet"
	contextTypeMMS      = "mms"
)

const (
	ofonoAttachInProgressError = "org.ofono.AttachInProgress"
	ofonoInProgressError       = "org.ofono.InProgress"
	ofonoNotAttachedError      = "org.ofono.Error.NotAttached"
)

type OfonoContext struct {
	ObjectPath dbus.ObjectPath
	Properties PropertiesType
}

type Modem struct {
	conn                   *dbus.Connection
	Modem                  dbus.ObjectPath
	PushAgent              *PushAgent
	identity               string
	IdentityAdded          chan string
	IdentityRemoved        chan string
	endWatch               chan bool
	PushInterfaceAvailable chan bool
	pushInterfaceAvailable bool
	online                 bool
	modemSignal, simSignal *dbus.SignalWatch
}

type ProxyInfo struct {
	Host string
	Port uint64
}

func (p ProxyInfo) String() string {
	return fmt.Sprintf("%s:%d", p.Host, p.Port)
}

func (oProp OfonoContext) String() string {
	var s string
	s += fmt.Sprintf("ObjectPath: %s\n", oProp.ObjectPath)
	for k, v := range oProp.Properties {
		s += fmt.Sprint("\t", k, ": ", v.Value, "\n")
	}
	return s
}

func NewModem(conn *dbus.Connection, objectPath dbus.ObjectPath) *Modem {
	return &Modem{
		conn:                   conn,
		Modem:                  objectPath,
		IdentityAdded:          make(chan string),
		IdentityRemoved:        make(chan string),
		PushInterfaceAvailable: make(chan bool),
		endWatch:               make(chan bool),
		PushAgent:              NewPushAgent(objectPath),
	}
}

func (modem *Modem) Init() (err error) {
	log.Printf("Initializing modem %s", modem.Modem)
	modem.modemSignal, err = connectToPropertySignal(modem.conn, modem.Modem, MODEM_INTERFACE)
	if err != nil {
		return err
	}

	modem.simSignal, err = connectToPropertySignal(modem.conn, modem.Modem, SIM_MANAGER_INTERFACE)
	if err != nil {
		return err
	}

	// the calling order here avoids race conditions
	go modem.watchStatus()
	modem.fetchExistingStatus()

	return nil
}

// fetchExistingStatus fetches key required for the modem to be considered operational
// from a push notification point of view
//
// status updates are fetched through dbus method calls
func (modem *Modem) fetchExistingStatus() {
	if v, err := modem.getProperty(MODEM_INTERFACE, "Interfaces"); err == nil {
		modem.updatePushInterfaceState(*v)
	} else {
		log.Print("Initial value couldn't be retrieved: ", err)
	}
	if v, err := modem.getProperty(MODEM_INTERFACE, "Online"); err == nil {
		modem.handleOnlineState(*v)
	} else {
		log.Print("Initial value couldn't be retrieved: ", err)
	}
	if v, err := modem.getProperty(SIM_MANAGER_INTERFACE, "SubscriberIdentity"); err == nil {
		modem.handleIdentity(*v)
	}
}

// watchStatus monitors key states required for the modem to be considered operational
// from a push notification point of view
//
// status updates are monitered by hooking up to dbus signals
func (modem *Modem) watchStatus() {
	var propName string
	var propValue dbus.Variant
watchloop:
	for {
		select {
		case <-modem.endWatch:
			log.Printf("Ending modem watch for %s", modem.Modem)
			break watchloop
		case msg, ok := <-modem.modemSignal.C:
			if !ok {
				modem.modemSignal.C = nil
				continue watchloop
			}
			if err := msg.Args(&propName, &propValue); err != nil {
				log.Printf("Cannot interpret Modem Property change: %s", err)
				continue watchloop
			}
			switch propName {
			case "Interfaces":
				modem.updatePushInterfaceState(propValue)
			case "Online":
				modem.handleOnlineState(propValue)
			default:
				continue watchloop
			}
		case msg, ok := <-modem.simSignal.C:
			if !ok {
				modem.simSignal.C = nil
				continue watchloop
			}
			if err := msg.Args(&propName, &propValue); err != nil {
				log.Printf("Cannot interpret Sim Property change: %s", err)
				continue watchloop
			}
			if propName != "SubscriberIdentity" {
				continue watchloop
			}
			modem.handleIdentity(propValue)
		}
	}
}

func (modem *Modem) handleOnlineState(propValue dbus.Variant) {
	origState := modem.online
	modem.online = reflect.ValueOf(propValue.Value).Bool()
	if modem.online != origState {
		log.Printf("Modem online: %t", modem.online)
	}
}

func (modem *Modem) handleIdentity(propValue dbus.Variant) {
	identity := reflect.ValueOf(propValue.Value).String()
	if identity == "" && modem.identity != "" {
		log.Printf("Identity before remove %s", modem.identity)

		modem.IdentityRemoved <- identity
		modem.identity = identity
	}
	log.Printf("Identity added %s", identity)
	if identity != "" && modem.identity == "" {
		modem.identity = identity
		modem.IdentityAdded <- identity
	}
}

func (modem *Modem) updatePushInterfaceState(interfaces dbus.Variant) {
	origState := modem.pushInterfaceAvailable
	availableInterfaces := reflect.ValueOf(interfaces.Value)
	for i := 0; i < availableInterfaces.Len(); i++ {
		interfaceName := reflect.ValueOf(availableInterfaces.Index(i).Interface().(string)).String()
		if interfaceName == PUSH_NOTIFICATION_INTERFACE {
			modem.pushInterfaceAvailable = true
			break
		}
	}
	if modem.pushInterfaceAvailable != origState {
		log.Printf("Push interface state: %t", modem.pushInterfaceAvailable)
		if modem.pushInterfaceAvailable {
			modem.PushInterfaceAvailable <- true
		} else if modem.PushAgent.Registered {
			modem.PushInterfaceAvailable <- false
		}
	}
}

var getOfonoProps = func(conn *dbus.Connection, objectPath dbus.ObjectPath, destination, iface, method string) (oProps []OfonoContext, err error) {
	obj := conn.Object(destination, objectPath)
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
func (modem *Modem) ActivateMMSContext(preferredContext dbus.ObjectPath) (OfonoContext, error) {
	contexts, err := modem.GetMMSContexts(preferredContext)
	if err != nil {
		return OfonoContext{}, err
	}
	for _, context := range contexts {
		if context.isActive() {
			return context, nil
		}
		if err := context.toggleActive(true, modem.conn); err == nil {
			return context, nil
		} else {
			log.Println("Failed to activate for", context.ObjectPath, ":", err)
		}
	}
	return OfonoContext{}, errors.New("no context available to activate")
}

//DeactivateMMSContext deactivates the context if it is of type mms
func (modem *Modem) DeactivateMMSContext(context OfonoContext) error {
	if context.isTypeInternet() {
		return nil
	}

	return context.toggleActive(false, modem.conn)
}

func activationErrorNeedsWait(err error) bool {
	if dbusErr, ok := err.(*dbus.Error); ok {
		return dbusErr.Name == ofonoInProgressError || dbusErr.Name == ofonoAttachInProgressError || dbusErr.Name == ofonoNotAttachedError
	}
	return false
}

func (context OfonoContext) toggleActive(state bool, conn *dbus.Connection) error {
	log.Println("Trying to set Active property to", state, "for context on", state, context.ObjectPath)
	obj := conn.Object("org.ofono", context.ObjectPath)
	for i := 0; i < 3; i++ {
		_, err := obj.Call(CONNECTION_CONTEXT_INTERFACE, "SetProperty", "Active", dbus.Variant{state})
		if err != nil {
			log.Printf("Cannot set Activate to %t (try %d/3) interface on %s: %s", state, i+1, context.ObjectPath, err)
			if activationErrorNeedsWait(err) {
				time.Sleep(2 * time.Second)
			}
		} else {
			return nil
		}
	}
	return errors.New("failed to activate context")
}

func (oContext OfonoContext) isTypeInternet() bool {
	if v, ok := oContext.Properties["Type"]; ok {
		return reflect.ValueOf(v.Value).String() == contextTypeInternet
	}
	return false
}

func (oContext OfonoContext) isTypeMMS() bool {
	if v, ok := oContext.Properties["Type"]; ok {
		return reflect.ValueOf(v.Value).String() == contextTypeMMS
	}
	return false
}

func (oContext OfonoContext) isActive() bool {
	return reflect.ValueOf(oContext.Properties["Active"].Value).Bool()
}

func (oContext OfonoContext) hasMessageCenter() bool {
	return oContext.messageCenter() != ""
}

func (oContext OfonoContext) messageCenter() string {
	if v, ok := oContext.Properties["MessageCenter"]; ok {
		return reflect.ValueOf(v.Value).String()
	}
	return ""
}

func (oContext OfonoContext) messageProxy() string {
	if v, ok := oContext.Properties["MessageProxy"]; ok {
		return reflect.ValueOf(v.Value).String()
	}
	return ""
}

func (oContext OfonoContext) name() string {
	if v, ok := oContext.Properties["Name"]; ok {
		return reflect.ValueOf(v.Value).String()
	}
	return ""
}

func (oContext OfonoContext) GetMessageCenter() (string, error) {
	if oContext.hasMessageCenter() {
		return oContext.messageCenter(), nil
	} else {
		return "", errors.New("context setting for the Message Center value is empty")
	}
}

func (oContext OfonoContext) GetProxy() (proxyInfo ProxyInfo, err error) {
	proxy := oContext.messageProxy()
	// we need to support empty proxies
	if proxy == "" {
		return proxyInfo, nil
	}
	if strings.HasPrefix(proxy, "http://") {
		proxy = proxy[len("http://"):]
	}
	var portString string
	proxyInfo.Host, portString, err = net.SplitHostPort(proxy)
	if err != nil {
		proxyInfo.Host = proxy
		proxyInfo.Port = 80
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
func (modem *Modem) GetMMSContexts(preferredContext dbus.ObjectPath) (mmsContexts []OfonoContext, err error) {
	contexts, err := getOfonoProps(modem.conn, modem.Modem, OFONO_SENDER, CONNECTION_MANAGER_INTERFACE, "GetContexts")
	if err != nil {
		return mmsContexts, err
	}

	for _, context := range contexts {
		if (context.isTypeInternet() && context.isActive() && context.hasMessageCenter()) || context.isTypeMMS() {
			if context.ObjectPath == preferredContext || context.isActive() {
				mmsContexts = append([]OfonoContext{context}, mmsContexts...)
			} else {
				mmsContexts = append(mmsContexts, context)
			}
		}
	}
	if len(mmsContexts) == 0 {
		log.Printf("non matching contexts:\n %+v", contexts)
		return mmsContexts, errors.New("No mms contexts found")
	}
	return mmsContexts, nil
}

func (modem *Modem) getProperty(interfaceName, propertyName string) (*dbus.Variant, error) {
	errorString := "Cannot retrieve %s from %s for %s: %s"
	rilObj := modem.conn.Object(OFONO_SENDER, modem.Modem)
	if reply, err := rilObj.Call(interfaceName, "GetProperties"); err == nil {
		var property PropertiesType
		if err := reply.Args(&property); err != nil {
			return nil, fmt.Errorf(errorString, propertyName, interfaceName, modem.Modem, err)
		}
		if v, ok := property[propertyName]; ok {
			return &v, nil
		}
		return nil, fmt.Errorf(errorString, propertyName, interfaceName, modem.Modem, "property not found")
	} else {
		return nil, fmt.Errorf(errorString, propertyName, interfaceName, modem.Modem, err)
	}
}

func (modem *Modem) Delete() {
	if modem.identity != "" {
		modem.IdentityRemoved <- modem.identity
	}
	modem.modemSignal.Cancel()
	modem.modemSignal.C = nil
	modem.simSignal.Cancel()
	modem.simSignal.C = nil
	modem.endWatch <- true
}
