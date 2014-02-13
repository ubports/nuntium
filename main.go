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
	"launchpad.net/nuntium/mms"
	"launchpad.net/go-dbus/v1"
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
		log.Fatal("Connection error:", err)
	}
	log.Print("Using dbus on ", conn.UniqueName)
	modems, err := NewModems(conn)
	if err != nil {
		log.Fatal("Could not add modems")
	}
	log.Print("Amount of modems found: ", len(modems))
	mmsChannel := make(chan *PushPDU)
	proxyChannel := make(chan ProxyInfo, 1)

	go func() {
		proxy := <-proxyChannel
		log.Print("Proxy set to ", proxy)
		for pushMsg := range mmsChannel {
			log.Print(pushMsg)
			dec := mms.NewDecoder(pushMsg.Data)
			mmsHdr := new(mms.MNotificationInd)
			if err := dec.Decode(mmsHdr); err != nil {
				log.Print("Unable to decode MMS Header", err)
			}
			log.Print(mmsHdr)
			//TODO send m-notifyresp.ind
			if filePath, err := mmsHdr.Download(proxy.Host, int32(proxy.Port), "", ""); err != nil {
				log.Print("Download issues ", err)
			} else {
				//TODO notify upper layer
				log.Print("downloaded ", filePath)
			}
		}
	}()

	//TODO refactor with new ofono work
	for i, _ := range modems {
		err := modems[i].GetContexts(conn, "mms")
		if err != nil {
			log.Print("Cannot get ofono context", err)
			continue
		}
		if len(modems[i].Contexts) == 0 {
			log.Print("No mms contexts found, no proxy setup")
		} else {
			log.Print("Getting context proxy")
			proxy, _ := modems[i].Contexts[0].GetProxy()
			proxyChannel <- proxy
		}
		if err := modems[i].RegisterAgent(conn, mmsChannel); err != nil {
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
