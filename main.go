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

	"launchpad.net/go-dbus/v1"
	"launchpad.net/nuntium/ofono"
	"launchpad.net/nuntium/telepathy"
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
	mmsManager, err := telepathy.NewMMSManager(connSession)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Using session bus on ", connSession.UniqueName)

	if conn, err = dbus.Connect(dbus.SystemBus); err != nil {
		log.Fatal("Connection error: ", err)
	}
	log.Print("Using system bus on ", conn.UniqueName)
	modems, err := ofono.NewModems(conn)
	if err != nil {
		log.Fatal("Could not add modems")
	}
	log.Print("Amount of modems found: ", len(modems))

	//TODO refactor with new ofono work
	for i, _ := range modems {
		mediator := NewMediator(modems[i])
		go mediator.init(mmsManager)
		if err := mediator.kickstart(); err != nil {
			log.Fatal(err)
		}
	}

	m := Mainloop{
		sigchan:  make(chan os.Signal, 1),
		termchan: make(chan int),
		Bindings: make(map[os.Signal]func())}

	m.Bindings[syscall.SIGHUP] = func() { m.Stop(); HupHandler() }
	m.Bindings[syscall.SIGINT] = func() { m.Stop(); IntHandler() }
	m.Start()
}
