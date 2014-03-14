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
	"launchpad.net/go-dbus/v1"
	"launchpad.net/nuntium/mms"
	"launchpad.net/nuntium/ofono"
	"log"
	"os"
	"syscall"
)

func main() {
	var (
		conn *dbus.Connection
		err  error
	)
	if conn, err = dbus.Connect(dbus.SystemBus); err != nil {
		log.Fatal("Connection error: ", err)
	}
	log.Print("Using dbus on ", conn.UniqueName)
	modems, err := ofono.NewModems(conn)
	if err != nil {
		log.Fatal("Could not add modems")
	}
	log.Print("Amount of modems found: ", len(modems))

	//TODO refactor with new ofono work
	for i, _ := range modems {
		_, err := modems[i].GetMMSContext(conn)
		if err != nil {
			log.Print("Cannot get ofono context: ", err)
			continue
		}
		pushChannel, err := modems[i].RegisterAgent(conn)
		go messageLoop(conn, pushChannel)
		if err != nil {
			log.Fatal(err)
		}
		defer modems[i].UnregisterAgent(conn)
	}

	m := Mainloop{
		sigchan:  make(chan os.Signal, 1),
		termchan: make(chan int),
		Bindings: make(map[os.Signal]func())}

	m.Bindings[syscall.SIGHUP] = func() { m.Stop(); HupHandler() }
	m.Bindings[syscall.SIGINT] = func() { m.Stop(); IntHandler() }
	m.Start()
}

func messageLoop(conn *dbus.Connection, mmsChannel chan *ofono.PushEvent) {
	for pushMsg := range mmsChannel {
		go func() {
			log.Print(pushMsg)
			dec := mms.NewDecoder(pushMsg.PDU.Data)
			mmsHdr := new(mms.MNotificationInd)
			if err := dec.Decode(mmsHdr); err != nil {
				log.Print("Unable to decode MMS Header: ", err)
			}
			mmsContext, err := pushMsg.Modem.ActivateMMSContext(conn)
			if err != nil {
				log.Print("Cannot activate ofono context: ", err)
				return
			}
			proxy, err := mmsContext.GetProxy()
			if err != nil {
				log.Print("Error retrieving proxy: ", err)
			}
			if filePath, err := mmsHdr.Download(proxy.Host, int32(proxy.Port), "", ""); err != nil {
				log.Print("Download issues: ", err)
			} else {
				//TODO notify upper layer
				log.Print("Downloaded ", filePath)
			}
			//TODO send m-notifyresp.ind
		}()
	}
}
