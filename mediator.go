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
	"log"
	"os"

	"launchpad.net/nuntium/mms"
	"launchpad.net/nuntium/ofono"
	"launchpad.net/nuntium/storage"
	"launchpad.net/nuntium/telepathy"
)

type Mediator struct {
	modem                 *ofono.Modem
	telepathyService      *telepathy.MMSService
	NewMNotificationInd   chan *mms.MNotificationInd
	NewMNotifyRespInd     chan *mms.MNotifyRespInd
	NewMRetrieveConf      chan *mms.MRetrieveConf
	NewMRetrieveConfFile  chan string
	NewMNotifyRespIndFile chan string
	terminate             chan bool
}

//TODO these vars need a configuration location managed by system settings or
//some UI accessible location.
//useDeliveryReports is set in ofono
var (
	deferredDownload   bool
	useDeliveryReports bool
)

func NewMediator(modem *ofono.Modem) *Mediator {
	mediator := &Mediator{modem: modem}
	mediator.NewMNotificationInd = make(chan *mms.MNotificationInd)
	mediator.NewMRetrieveConf = make(chan *mms.MRetrieveConf)
	mediator.NewMRetrieveConfFile = make(chan string)
	mediator.NewMNotifyRespInd = make(chan *mms.MNotifyRespInd)
	mediator.NewMNotifyRespIndFile = make(chan string)
	mediator.terminate = make(chan bool)
	return mediator
}

func (mediator *Mediator) Delete() {
	mediator.terminate <- mediator.telepathyService == nil
}

func (mediator *Mediator) init(mmsManager *telepathy.MMSManager) {
mediatorLoop:
	for {
		select {
		case push, ok := <-mediator.modem.PushAgent.Push:
			if !ok {
				log.Print("PushChannel is closed")
				continue
			}
			go mediator.handleMNotificationInd(push)
		case mNotificationInd := <-mediator.NewMNotificationInd:
			if deferredDownload {
				go mediator.handleDeferredDownload(mNotificationInd)
			} else {
				go mediator.getMRetrieveConf(mNotificationInd)
			}
		case mRetrieveConfFilePath := <-mediator.NewMRetrieveConfFile:
			go mediator.handleMRetrieveConf(mRetrieveConfFilePath)
		case mRetrieveConf := <-mediator.NewMRetrieveConf:
			go mediator.handleRetrieved(mRetrieveConf)
		case mNotifyRespInd := <-mediator.NewMNotifyRespInd:
			go mediator.handleMNotifyRespInd(mNotifyRespInd)
		case mNotifyRespIndFilePath := <-mediator.NewMNotifyRespIndFile:
			go mediator.sendMNotifyRespInd(mNotifyRespIndFilePath)
		case id := <-mediator.modem.IdentityAdded:
			var err error
			mediator.telepathyService, err = mmsManager.AddService(id, useDeliveryReports)
			if err != nil {
				log.Fatal(err)
			}
		case id := <-mediator.modem.IdentityRemoved:
			err := mmsManager.RemoveService(id)
			if err != nil {
				log.Fatal(err)
			}
			mediator.telepathyService = nil
		case ok := <-mediator.modem.PushInterfaceAvailable:
			if ok {
				if err := mediator.modem.PushAgent.Register(); err != nil {
					log.Fatal(err)
				}
			} else {
				if err := mediator.modem.PushAgent.Unregister(); err != nil {
					log.Fatal(err)
				}
			}
		case terminate := <-mediator.terminate:
			/*
				close(mediator.terminate)
				close(mediator.NewMNotificationInd)
				close(mediator.NewMRetrieveConf)
				close(mediator.NewMRetrieveConfFile)
				close(mediator.NewMNotifyRespInd)
				close(mediator.NewMNotifyRespIndFile)
			*/
			if terminate {
				break mediatorLoop
			}
		}
	}
	log.Print("Ending mediator instance loop for modem")
}

func (mediator *Mediator) handleMNotificationInd(pushMsg *ofono.PushPDU) {
	if pushMsg == nil {
		log.Print("Received nil push")
		return
	}
	dec := mms.NewDecoder(pushMsg.Data)
	mNotificationInd := mms.NewMNotificationInd()
	if err := dec.Decode(mNotificationInd); err != nil {
		log.Print("Unable to decode m-notification.ind: ", err)
		return
	}
	storage.Create(mNotificationInd.UUID, mNotificationInd.ContentLocation)
	mediator.NewMNotificationInd <- mNotificationInd
}

func (mediator *Mediator) handleDeferredDownload(mNotificationInd *mms.MNotificationInd) {
	//TODO send MessageAdded with status="deferred" and mNotificationInd relevant headers
}

func (mediator *Mediator) getMRetrieveConf(mNotificationInd *mms.MNotificationInd) {
	mmsContext, err := mediator.modem.ActivateMMSContext()
	if err != nil {
		log.Print("Cannot activate ofono context: ", err)
		return
	}
	proxy, err := mmsContext.GetProxy()
	if err != nil {
		log.Print("Error retrieving proxy: ", err)
		return
	}
	if filePath, err := mNotificationInd.DownloadContent(proxy.Host, int32(proxy.Port)); err != nil {
		//TODO telepathy service signal the download error
		log.Print("Download issues: ", err)
		return
	} else {
		storage.UpdateDownloaded(mNotificationInd.UUID, filePath)
	}
	mediator.NewMRetrieveConfFile <- mNotificationInd.UUID
}

func (mediator *Mediator) handleMRetrieveConf(uuid string) {
	var filePath string
	if f, err := storage.GetMMS(uuid); err == nil {
		filePath = f
	} else {
		log.Print("Unable to retrieve MMS: ", err)
		return
	}
	mmsData, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Print("Issues while reading from downloaded file: ", err)
		return
	}
	mRetrieveConf := mms.NewMRetrieveConf(uuid)
	dec := mms.NewDecoder(mmsData)
	if err := dec.Decode(mRetrieveConf); err != nil {
		log.Print("Unable to decode m-retrieve.conf: ", err)
		return
	}
	mediator.NewMRetrieveConf <- mRetrieveConf
	if mediator.telepathyService != nil {
		if err := mediator.telepathyService.MessageAdded(mRetrieveConf); err != nil {
			log.Println("Cannot notify telepathy-ofono about new message", err)
		}
	} else {
		log.Print("Not sending recently retrieved message")
	}
}

func (mediator *Mediator) handleRetrieved(mRetrieveConf *mms.MRetrieveConf) {
	//TODO chann for send m-notifyresp.ind
	mNotifyRespInd := mRetrieveConf.NewMNotifyRespInd(useDeliveryReports)
	if err := storage.UpdateRetrieved(mNotifyRespInd.UUID); err != nil {
		log.Print("Can't update mms status: ", err)
		return
	}
	mediator.NewMNotifyRespInd <- mNotifyRespInd
}

func (mediator *Mediator) handleMNotifyRespInd(mNotifyRespInd *mms.MNotifyRespInd) {
	f, err := storage.CreateResponseFile(mNotifyRespInd.UUID)
	if err != nil {
		log.Print("Unable to create m-notifyresp.ind file for ", mNotifyRespInd.UUID)
		return
	}
	defer f.Close()
	enc := mms.NewEncoder(f)
	if err := enc.Encode(mNotifyRespInd); err != nil {
		log.Print("Unable to encode m-notifyresp.ind for ", mNotifyRespInd.UUID)
		return
	}
	filePath := f.Name()
	log.Printf("Created %s to handle m-notifyresp.ind for %s", filePath, mNotifyRespInd.UUID)
	mediator.NewMNotifyRespIndFile <- filePath
}

func (mediator *Mediator) sendMNotifyRespInd(mNotifyRespIndFile string) {
	defer os.Remove(mNotifyRespIndFile)
	if err := mediator.uploadFile(mNotifyRespIndFile); err != nil {
		log.Printf("Cannot upload m-notifyresp.ind encoded file %s to message center: %s", mNotifyRespIndFile, err)
	}
}

func (mediator *Mediator) uploadFile(filePath string) error {
	mmsContext, err := mediator.modem.ActivateMMSContext()
	if err != nil {
		log.Print("Cannot activate ofono context: ", err)
		return err
	}
	proxy, err := mmsContext.GetProxy()
	if err != nil {
		log.Print("Error retrieving proxy: ", err)
		return err
	}
	msc, err := mmsContext.GetMessageCenter()
	if err != nil {
		log.Print("Error retrieving Message Center: ", err)
		return err
	}
	if err := mms.Upload(filePath, msc, proxy.Host, int32(proxy.Port)); err != nil {
		log.Printf("Cannot upload %s: %s", filePath, err)
		return err
	}
	return nil
}
