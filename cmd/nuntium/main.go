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

package main

import (
	"log"
	"os"
	"syscall"

	"github.com/ubuntu-phonedations/nuntium/ofono"
	"github.com/ubuntu-phonedations/nuntium/telepathy"
	"launchpad.net/go-dbus/v1"
)

func main() {
	var (
		conn        *dbus.Connection
		connSession *dbus.Connection
		err         error
	)
	if connSession, err = dbus.Connect(dbus.SessionBus); err != nil {
		log.Fatal("Connection error: ", err)
	}
	log.Print("Using session bus on ", connSession.UniqueName)

	mmsManager, err := telepathy.NewMMSManager(connSession)
	if err != nil {
		log.Fatal(err)
	}

	if conn, err = dbus.Connect(dbus.SystemBus); err != nil {
		log.Fatal("Connection error: ", err)
	}
	log.Print("Using system bus on ", conn.UniqueName)

	modemManager := ofono.NewModemManager(conn)
	mediators := make(map[dbus.ObjectPath]*Mediator)
	go func() {
		for {
			select {
			case modem := <-modemManager.ModemAdded:
				mediators[modem.Modem] = NewMediator(modem)
				go mediators[modem.Modem].init(mmsManager)
				if err := modem.Init(); err != nil {
					log.Printf("Cannot initialize modem %s", modem.Modem)
				}
			case modem := <-modemManager.ModemRemoved:
				mediators[modem.Modem].Delete()
			}
		}
	}()

	if err := modemManager.Init(); err != nil {
		log.Fatal(err)
	}

	m := Mainloop{
		sigchan:  make(chan os.Signal, 1),
		termchan: make(chan int),
		Bindings: make(map[os.Signal]func())}

	m.Bindings[syscall.SIGHUP] = func() { m.Stop(); HupHandler() }
	m.Bindings[syscall.SIGINT] = func() { m.Stop(); IntHandler() }
	m.Start()
}
