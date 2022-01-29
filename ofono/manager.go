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
	"launchpad.net/go-dbus"
	"log"
)

type Modems map[dbus.ObjectPath]*Modem

type ModemManager struct {
	ModemAdded   chan (*Modem)
	ModemRemoved chan (*Modem)
	modems       Modems
	conn         *dbus.Connection
}

func NewModemManager(conn *dbus.Connection) *ModemManager {
	return &ModemManager{
		conn:         conn,
		ModemAdded:   make(chan *Modem),
		ModemRemoved: make(chan *Modem),
		modems:       make(Modems),
	}
}

func (mm *ModemManager) Init() error {
	//Use a different connection for the modem signals to avoid go-dbus blocking issues
	conn, err := dbus.Connect(dbus.SystemBus)
	if err != nil {
		return err
	}

	modemAddedSignal, err := connectToSignal(conn, "/", OFONO_MANAGER_INTERFACE, "ModemAdded")
	if err != nil {
		return err
	}
	modemRemovedSignal, err := connectToSignal(conn, "/", OFONO_MANAGER_INTERFACE, "ModemRemoved")
	if err != nil {
		return err
	}
	go mm.watchModems(modemAddedSignal, modemRemovedSignal)

	//Check for existing modems
	modemPaths, err := getModems(conn)
	if err != nil {
		log.Print("Cannot preemptively add modems: ", err)
	} else {
		for _, objectPath := range modemPaths {
			mm.addModem(objectPath)
		}
	}
	return nil
}

func (mm *ModemManager) watchModems(modemAdded, modemRemoved *dbus.SignalWatch) {
	for {
		var objectPath dbus.ObjectPath
		select {
		case m := <-modemAdded.C:
			var signalProps PropertiesType
			if err := m.Args(&objectPath, &signalProps); err != nil {
				log.Print(err)
				continue
			}
			mm.addModem(objectPath)
		case m := <-modemRemoved.C:
			if err := m.Args(&objectPath); err != nil {
				log.Print(err)
				continue
			}
			mm.removeModem(objectPath)
		}
	}
}

func (mm *ModemManager) addModem(objectPath dbus.ObjectPath) {
	if modem, ok := mm.modems[objectPath]; ok {
		log.Printf("Need to delete stale modem instance %s", modem.Modem)
		modem.Delete()
		delete(mm.modems, objectPath)
	}
	mm.modems[objectPath] = NewModem(mm.conn, objectPath)
	mm.ModemAdded <- mm.modems[objectPath]
}

func (mm *ModemManager) removeModem(objectPath dbus.ObjectPath) {
	if modem, ok := mm.modems[objectPath]; ok {
		mm.ModemRemoved <- mm.modems[objectPath]
		log.Printf("Deleting modem instance %s", modem.Modem)
		modem.Delete()
		delete(mm.modems, objectPath)
	} else {
		log.Printf("Cannot satisfy request to remove modem %s as it does not exist", objectPath)
	}
}
