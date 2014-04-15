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
	"io/ioutil"
	"launchpad.net/go-dbus/v1"
	"launchpad.net/nuntium/mms"
	"launchpad.net/nuntium/ofono"
	"launchpad.net/nuntium/telepathy"
	"log"
	"os"
	"syscall"
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
		go func() {
			var telepathyService *telepathy.MMSService
			for {
				select {
				case id := <-modems[i].IdentityAdded:
					telepathyService, err = mmsManager.AddService(id, false)
					if err != nil {
						log.Fatal(err)
					}
				case id := <-modems[i].IdentityRemoved:
					err := mmsManager.RemoveService(id)
					if err != nil {
						log.Fatal(err)
					}
					telepathyService = nil
				case <-modems[i].ReadySignal:
					if err := modems[i].RegisterAgent(conn); err != nil {
						log.Fatal("Error while registering agent: ", err)
					}
				case pushMsg := <-modems[i].PushChannel:
					go processMessage(conn, pushMsg, telepathyService)
				}
			}
		}()
		if err := modems[i].WatchPushInterface(conn); err != nil {
			log.Fatal(err)
		}
		if err := modems[i].GetIdentity(conn); err != nil {
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

func processMessage(conn *dbus.Connection, pushMsg *ofono.PushEvent, telepathyService *telepathy.MMSService) {
	if pushMsg == nil {
		return
	}
	log.Print(pushMsg)
	dec := mms.NewDecoder(pushMsg.PDU.Data)
	mmsIndHdr := mms.NewMNotificationInd()
	if err := dec.Decode(mmsIndHdr); err != nil {
		log.Print("Unable to decode m-notification.ind: ", err)
		return
	}
	mmsContext, err := pushMsg.Modem.ActivateMMSContext(conn)
	if err != nil {
		log.Print("Cannot activate ofono context: ", err)
		return
	}
	proxy, err := mmsContext.GetProxy()
	if err != nil {
		log.Print("Error retrieving proxy: ", err)
		return
	}
	filePath, err := mmsIndHdr.DownloadContent(proxy.Host, int32(proxy.Port))
	if err != nil {
		log.Print("Download issues: ", err)
		return
	}
	log.Print("Downloaded ", filePath)
	mmsData, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Print("Issues while reading from downloaded file: ", err)
		return
	}
	mmsRetConfHdr := mms.NewMRetrieveConf(filePath)
	dec = mms.NewDecoder(mmsData)
	if err := dec.Decode(mmsRetConfHdr); err != nil {
		log.Print("Unable to decode m-retrieve.conf: ", err)
		return
	}
	//TODO send m-notifyresp.ind
	if telepathyService != nil {
		telepathyService.MessageAdded(mmsRetConfHdr)
	} else {
		log.Print("Not sending recently retrieved message")
	}
}
