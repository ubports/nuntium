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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"sync"
	"time"

	"github.com/ubports/nuntium/mms"
	"github.com/ubports/nuntium/ofono"
	"github.com/ubports/nuntium/storage"
	"github.com/ubports/nuntium/telepathy"
	"launchpad.net/go-dbus/v1"
)

type Mediator struct {
	modem               *ofono.Modem
	telepathyService    *telepathy.MMSService
	NewMNotificationInd chan *mms.MNotificationInd
	NewMSendReq         chan *mms.MSendReq
	NewMSendReqFile     chan struct{ filePath, uuid string }
	outMessage          chan *telepathy.OutgoingMessage
	terminate           chan bool
	contextLock         sync.Mutex
	undownloaded        map[string]string // (transactionId: UUID) Messages with download error, which was successfully communicated to telepathyService.
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
	mediator.NewMSendReq = make(chan *mms.MSendReq)
	mediator.NewMSendReqFile = make(chan struct{ filePath, uuid string })
	mediator.outMessage = make(chan *telepathy.OutgoingMessage)
	mediator.terminate = make(chan bool)
	mediator.undownloaded = make(map[string]string)
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
			if !mmsEnabled() {
				log.Print("MMS is disabled")
				continue
			}
			go mediator.handleMNotificationInd(push, mediator.modem.Identity())
		case mNotificationInd := <-mediator.NewMNotificationInd:
			if deferredDownload {
				go mediator.handleDeferredDownload(mNotificationInd)
			} else {
				go mediator.getMRetrieveConf(mNotificationInd)
			}
		case msg := <-mediator.outMessage:
			go mediator.handleOutgoingMessage(msg)
		case mSendReq := <-mediator.NewMSendReq:
			go mediator.handleMSendReq(mSendReq)
		case mSendReqFile := <-mediator.NewMSendReqFile:
			go mediator.sendMSendReq(mSendReqFile.filePath, mSendReqFile.uuid)
		case id := <-mediator.modem.IdentityAdded:
			var err error
			mediator.telepathyService, err = mmsManager.AddService(id, mediator.modem.Modem, mediator.outMessage, useDeliveryReports, mediator.NewMNotificationInd)
			if err != nil {
				log.Fatal(err)
			}

			mediator.initializeMessages(id)
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
				close(mediator.outMessage)
				close(mediator.NewMNotificationInd)
				close(mediator.NewMRetrieveConf)
				close(mediator.NewMRetrieveConfFile)
				close(mediator.NewMSendReq)
				close(mediator.NewMSendReqFile)
			*/
			if terminate {
				break mediatorLoop
			}
		}
	}
	log.Print("Ending mediator instance loop for modem")
}

func (mediator *Mediator) handleMNotificationInd(pushMsg *ofono.PushPDU, modemId string) {
	if pushMsg == nil {
		log.Print("Received nil push")
		return
	}

	dec := mms.NewDecoder(pushMsg.Data)
	mNotificationInd := mms.NewMNotificationInd(time.Now())
	if err := dec.Decode(mNotificationInd); err != nil {
		log.Println("Unable to decode m-notification.ind: ", err, "with log", dec.GetLog())
		return
	}

	// Set received date to first push occurrence, if this is not a first time this transaction ID occurred.
	if mNotificationInd.TransactionId != "" {
		if uuid, ok := mediator.undownloaded[mNotificationInd.TransactionId]; ok {
			log.Printf("Pushed transaction ID (%s) is in undownloaded pointing to UUID: %s", mNotificationInd.TransactionId, uuid)
			if st, err := storage.GetMMSState(uuid); err == nil {
				log.Printf("jezek - mmsState: %v", st)
				if st.MNotificationInd != nil {
					log.Printf("Changing recieved date to the first push date: %v", st.MNotificationInd.Received)
					mNotificationInd.Received = st.MNotificationInd.Received
				} else {
					log.Printf("Error, no MNotificationInd in loaded mmsState for UUID %s", uuid)
				}
			} else {
				log.Printf("Error, can't load mmsState for UUID %s: %v", uuid, err)
			}
		}
	}

	storage.Create(modemId, mNotificationInd)
	mediator.NewMNotificationInd <- mNotificationInd
}

func (mediator *Mediator) handleDeferredDownload(mNotificationInd *mms.MNotificationInd) {
	//TODO send MessageAdded with status="deferred" and mNotificationInd relevant headers
	//
	//Reading:
	//	http://www.openmobilealliance.org/release/MMS/V1_1-20021104-C/OMA-WAP-MMS-ENC-V1_1-20021030-C.pdf - no deferred instructions, just mentions.
	//	https://dl.cdn-anritsu.com/en-gb/test-measurement/files/Technical-Notes/White-Paper/MC-MMS_Signaling_Examples_and_KPI_Calculations-WP-1_0.pdf - no defered instructions, just mentions.
	//	https://developer.brewmp.com/resources/tech-guides/multimedia-messaging-service-mms-technology-guide/mms-protocol-overview/mms-fe/receiving-mms-message - instructions on how to deffer.
	//	https://www.slideshare.net/glebodic/mobile-messaging-part-5-76-mms-arch-and-transactions-reduced - has deferred instructions
	//
	//Notes:
	//	If the application chooses to download the message at a later point in time, then an M-NotifyResp.ind is sent with the deferred flag set to acknowledge the receipt notification and notify that message download is deferred.

}

type standartizedError struct {
	error
	code string
}

func (e standartizedError) Code() string { return e.code }

type downloadError struct {
	standartizedError
}

func (e downloadError) AllowRedownload() bool { return true }

func (mediator *Mediator) getMRetrieveConf(mNotificationInd *mms.MNotificationInd) {
	mediator.contextLock.Lock()
	defer mediator.contextLock.Unlock()

	var proxy ofono.ProxyInfo
	var mmsContext ofono.OfonoContext
	if mNotificationInd.IsLocal() {
		log.Print("This is a local test, skipping context activation and proxy settings")
	} else {
		//TODO:jezek - encapsulate into deactivateFunc, err := activateMMSContext(...); defer deactivateFunc();
		var err error
		preferredContext, _ := mediator.telepathyService.GetPreferredContext()
		mmsContext, err = mediator.modem.ActivateMMSContext(preferredContext)
		if err != nil {
			log.Print("Cannot activate ofono context: ", err)
			mediator.handleMRetrieveConfDownloadError(mNotificationInd, downloadError{standartizedError{err, "x-ubports-nuntium-mms-error-activate-context"}})
			return
		}
		defer func() {
			if err := mediator.modem.DeactivateMMSContext(mmsContext); err != nil {
				log.Println("Issues while deactivating context:", err)
			}
		}()

		if err := mediator.telepathyService.SetPreferredContext(mmsContext.ObjectPath); err != nil {
			log.Println("Unable to store the preferred context for MMS:", err)
		}
		proxy, err = mmsContext.GetProxy()
		if err != nil {
			log.Print("Error retrieving proxy: ", err)
			mediator.handleMRetrieveConfDownloadError(mNotificationInd, downloadError{standartizedError{err, "x-ubports-nuntium-mms-error-get-proxy"}})
			return
		}
	}

	// Download message content.
	if filePath, err := mNotificationInd.DownloadContent(proxy.Host, int32(proxy.Port)); err != nil {
		log.Print("Download issues: ", err)
		mediator.handleMRetrieveConfDownloadError(mNotificationInd, downloadError{standartizedError{err, "x-ubports-nuntium-mms-error-download-content"}})
		return
	} else {
		if err := storage.UpdateDownloaded(mNotificationInd.UUID, filePath); err != nil {
			log.Println("Error updating storage (UpdateDownloaded): ", err)
			mediator.handleMRetrieveConfDownloadError(mNotificationInd, standartizedError{err, "x-ubports-nuntium-mms-error-storage"})
			return
		}
	}

	//TODO:jezek - split handleMRetrieveConf into getMRetrieveConf & handle...
	// Forward message to telepathy-ofono service.
	mRetrieveConf, err := mediator.handleMRetrieveConf(mNotificationInd)
	if err != nil {
		log.Printf("Handling MRetrieveConf error: %v", err)
		mediator.handleMRetrieveConfDownloadError(mNotificationInd, standartizedError{err, "x-ubports-nuntium-mms-error-forward"})
		return
	}
	if err := storage.UpdateReceived(mRetrieveConf.UUID); err != nil {
		log.Println("Error updating storage (UpdateRetrieved): ", err)
		return
	}

	// Notify MMS service about successful download.
	mNotifyRespInd := mRetrieveConf.NewMNotifyRespInd(useDeliveryReports)
	if !mNotificationInd.IsLocal() {
		// TODO deferred case
		filePath := mediator.handleMNotifyRespInd(mNotifyRespInd)
		if filePath == "" {
			return
		}
		mediator.sendMNotifyRespInd(filePath, &mmsContext)
	} else {
		log.Print("This is a local test, skipping m-notifyresp.ind")
	}
	//TODO:jezek - Add storage states to docs graph file docs/assets/receiving_success_deferral_disabled.msc
	if err := storage.UpdateResponded(mNotifyRespInd.UUID); err != nil {
		log.Println("Error updating storage (UpdateResponded): ", err)
		return
	}

	delete(mediator.undownloaded, mNotificationInd.TransactionId)
}

// Communicates the download error "err" of mNotificationInd to telepathy-ofono service.
// Some operators repeatedly push mNotificationInd with the same transaction id, if download not acknowledged by mNotifyRespInd. So we have to make sure, to communicate the download error just once.
func (mediator *Mediator) handleMRetrieveConfDownloadError(mNotificationInd *mms.MNotificationInd, err error) {
	_, inUndownloaded := mediator.undownloaded[mNotificationInd.TransactionId]
	if mNotificationInd.RedownloadOfUUID == "" && inUndownloaded && mNotificationInd.TransactionId != "" {
		// This download error "err" happened not after redownload and not after first download fail of pushed mNotificationInd with the same transaction id.
		log.Printf("Download error for MNotificationInd with TransactionId: \"%s\" was already communicated by UUID: \"%s\"", mNotificationInd.TransactionId, mediator.undownloaded[mNotificationInd.TransactionId])
		// Delete the message from storage.
		if err := storage.Destroy(mNotificationInd.UUID); err != nil {
			log.Printf("Error removing message %s from storage: %v", mNotificationInd.UUID, err)
			return
		}
		log.Printf("Message %s was removed from storage", mNotificationInd.UUID)
		return
	}

	// Error occurred after redownload requested or this is the first time the same download error for TransactionId occurred (or is empty, but this shouldn't happen)
	// Send error message to telepathy-ofono service.
	if addErr := mediator.telepathyService.IncomingMessageFailAdded(mNotificationInd, err); addErr != nil {
		// Couldn't inform telepathy-ofono about download fail.
		//TODO Store into storage & re-try on initilization.
		log.Printf("Sending download error message to telepathy-ofono has failed with error: %v", addErr)
		return
	}

	if mNotificationInd.TransactionId != "" {
		// Mark that some download error for TransactionId occurred and was already successfully communicated to telepathy-ofono service.
		mediator.undownloaded[mNotificationInd.TransactionId] = mNotificationInd.UUID
	}
}

func (mediator *Mediator) handleMRetrieveConf(mNotificationInd *mms.MNotificationInd) (*mms.MRetrieveConf, error) {
	var filePath string
	if f, err := storage.GetMMS(mNotificationInd.UUID); err == nil {
		filePath = f
	} else {
		return nil, fmt.Errorf("unable to retrieve MMS: %s", err)
	}

	mmsData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("issues while reading from downloaded file: %s", err)
	}

	mRetrieveConf := mms.NewMRetrieveConf(mNotificationInd.UUID)
	dec := mms.NewDecoder(mmsData)
	if err := dec.Decode(mRetrieveConf); err != nil {
		return nil, fmt.Errorf("unable to decode m-retrieve.conf: %s with log %s", err, dec.GetLog())
	}

	if mediator.telepathyService == nil {
		return nil, fmt.Errorf("no telepathy-ofono service found")
	}

	// Check if there was some download error communicated for TransactionId before and no redownload was triggered (on redownload request, RedownloadOfUUID is filled and listener is stopped automatically).
	if uuid, ok := mediator.undownloaded[mNotificationInd.TransactionId]; mNotificationInd.RedownloadOfUUID == "" && ok {
		// There was an error message communicated to telepathy before, mark it to delete it by telepathy when communicating this message.
		mNotificationInd.RedownloadOfUUID = uuid
		// Before return, close listener for redownload request.
		defer func() {
			if err := mediator.telepathyService.MessageRemoved(mediator.telepathyService.GenMessagePath(uuid)); err != nil {
				// Just log possible errors.
				log.Printf("Error closing meesage %s handlers: %v", uuid, err)
			}
		}()
	}

	// Forward message to telepathy-ofono service.
	if err := mediator.telepathyService.IncomingMessageAdded(mRetrieveConf, mNotificationInd); err != nil {
		return nil, fmt.Errorf("cannot notify telepathy-ofono about new message: %v", err)
	}

	return mRetrieveConf, nil
}

func (mediator *Mediator) handleMNotifyRespInd(mNotifyRespInd *mms.MNotifyRespInd) string {
	f, err := storage.CreateResponseFile(mNotifyRespInd.UUID)
	if err != nil {
		log.Print("Unable to create m-notifyresp.ind file for ", mNotifyRespInd.UUID)
		return ""
	}
	enc := mms.NewEncoder(f)
	if err := enc.Encode(mNotifyRespInd); err != nil {
		log.Print("Unable to encode m-notifyresp.ind for ", mNotifyRespInd.UUID)
		f.Close()
		return ""
	}
	filePath := f.Name()
	if err := f.Sync(); err != nil {
		log.Print("Error while syncing", f.Name(), ": ", err)
		return ""
	}
	if err := f.Close(); err != nil {
		log.Print("Error while closing", f.Name(), ": ", err)
		return ""
	}
	log.Printf("Created %s to handle m-notifyresp.ind for %s", filePath, mNotifyRespInd.UUID)
	return filePath
}

func (mediator *Mediator) sendMNotifyRespInd(filePath string, mmsContext *ofono.OfonoContext) {
	defer os.Remove(filePath)

	proxy, err := mmsContext.GetProxy()
	if err != nil {
		log.Println("Cannot retrieve MMS proxy setting", err)
		return
	}
	msc, err := mmsContext.GetMessageCenter()
	if err != nil {
		log.Println("Cannot retrieve MMSC setting", err)
		return
	}

	if _, err := mms.Upload(filePath, msc, proxy.Host, int32(proxy.Port)); err != nil {
		log.Printf("Cannot upload m-notifyresp.ind encoded file %s to message center: %s", filePath, err)
	}
}

func (mediator *Mediator) handleOutgoingMessage(msg *telepathy.OutgoingMessage) {
	var cts []*mms.Attachment
	for _, att := range msg.Attachments {
		ct, err := mms.NewAttachment(att.Id, att.ContentType, att.FilePath)
		if err != nil {
			log.Print(err)
			//TODO reply to telepathy ofono with an error
			return
		}
		cts = append(cts, ct)
	}
	mSendReq := mms.NewMSendReq(msg.Recipients, cts, useDeliveryReports)
	if _, err := mediator.telepathyService.ReplySendMessage(msg.Reply, mSendReq.UUID); err != nil {
		log.Print(err)
		return
	}
	mediator.NewMSendReq <- mSendReq
}

func (mediator *Mediator) handleMSendReq(mSendReq *mms.MSendReq) {
	log.Print("Encoding M-Send.Req")
	//TODO:jezek - storage is created, but it seems it is not deleted anywhere. Ensure deletion.
	//TODO:jezek - on initialize, handle undeleted send messages (also add modem id and on init delete old stored messages).
	f, err := storage.CreateSendFile(mSendReq.UUID)
	if err != nil {
		log.Print("Unable to create m-send.req file for ", mSendReq.UUID)
		return
	}
	defer f.Close()
	enc := mms.NewEncoder(f)
	if err := enc.Encode(mSendReq); err != nil {
		log.Print("Unable to encode m-send.req for ", mSendReq.UUID)
		if err := mediator.telepathyService.MessageStatusChanged(mSendReq.UUID, telepathy.PERMANENT_ERROR); err != nil {
			log.Println(err)
		}
		f.Close()
		return
	}
	filePath := f.Name()
	if err := f.Sync(); err != nil {
		log.Print("Error while syncing", f.Name(), ": ", err)
		return
	}
	if err := f.Close(); err != nil {
		log.Print("Error while closing", f.Name(), ": ", err)
		return
	}
	log.Printf("Created %s to handle m-send.req for %s", filePath, mSendReq.UUID)
	mediator.sendMSendReq(filePath, mSendReq.UUID)
}

func (mediator *Mediator) sendMSendReq(mSendReqFile, uuid string) {
	defer os.Remove(mSendReqFile)
	defer mediator.telepathyService.MessageDestroy(uuid)
	mSendConfFile, err := mediator.uploadFile(mSendReqFile)
	if err != nil {
		if err := mediator.telepathyService.MessageStatusChanged(uuid, telepathy.TRANSIENT_ERROR); err != nil {
			log.Println(err)
		}
		log.Printf("Cannot upload m-send.req encoded file %s to message center: %s", mSendReqFile, err)
		return
	}

	defer os.Remove(mSendConfFile)
	mSendConf, err := parseMSendConfFile(mSendConfFile)
	if err != nil {
		log.Println("Error while decoding m-send.conf:", err)
		if err := mediator.telepathyService.MessageStatusChanged(uuid, telepathy.TRANSIENT_ERROR); err != nil {
			log.Println(err)
		}
		return
	}

	log.Println("m-send.conf ResponseStatus for", uuid, "is", mSendConf.ResponseStatus)
	var status string
	switch mSendConf.Status() {
	case nil:
		status = telepathy.SENT
	case mms.ErrPermanent:
		status = telepathy.PERMANENT_ERROR
	case mms.ErrTransient:
		status = telepathy.TRANSIENT_ERROR
	}
	if err := mediator.telepathyService.MessageStatusChanged(uuid, status); err != nil {
		log.Println(err)
	}
}

func parseMSendConfFile(mSendConfFile string) (*mms.MSendConf, error) {
	b, err := ioutil.ReadFile(mSendConfFile)
	if err != nil {
		return nil, err
	}

	mSendConf := mms.NewMSendConf()

	dec := mms.NewDecoder(b)
	if err := dec.Decode(mSendConf); err != nil {
		return nil, err
	}
	return mSendConf, nil
}

func (mediator *Mediator) uploadFile(filePath string) (string, error) {
	mediator.contextLock.Lock()
	defer mediator.contextLock.Unlock()

	preferredContext, _ := mediator.telepathyService.GetPreferredContext()
	mmsContext, err := mediator.modem.ActivateMMSContext(preferredContext)
	if err != nil {
		return "", err
	}
	if err := mediator.telepathyService.SetPreferredContext(mmsContext.ObjectPath); err != nil {
		log.Println("Unable to store the preferred context for MMS:", err)
	}
	defer func() {
		if err := mediator.modem.DeactivateMMSContext(mmsContext); err != nil {
			log.Println("Issues while deactivating context:", err)
		}
	}()

	proxy, err := mmsContext.GetProxy()
	if err != nil {
		return "", err
	}
	msc, err := mmsContext.GetMessageCenter()
	if err != nil {
		return "", err
	}
	mSendRespFile, uploadErr := mms.Upload(filePath, msc, proxy.Host, int32(proxy.Port))

	return mSendRespFile, uploadErr
}

// By default this method returns true, unless it is strictly requested to disable.
func mmsEnabled() bool {
	conn, err := dbus.Connect(dbus.SystemBus)
	if err != nil {
		log.Printf("mmsEnabled: connecting to dbus failed: %v", err)
		return true
	}

	usr, err := user.Current()
	if err != nil {
		log.Printf("mmsEnabled: getting user failed: %v", err)
		return true
	}
	activeUser := dbus.ObjectPath("/org/freedesktop/Accounts/User" + usr.Uid)

	obj := conn.Object("org.freedesktop.Accounts", activeUser)
	reply, err := obj.Call("org.freedesktop.DBus.Properties", "Get", "com.ubuntu.touch.AccountsService.Phone", "MmsEnabled")
	if err != nil || reply.Type == dbus.TypeError {
		log.Printf("mmsEnabled: failed to get mms option: %v", err)
		return true
	}

	mms := dbus.Variant{true}
	if err := reply.Args(&mms); err != nil {
		log.Printf("mmsEnabled: failed to get mms option reply: %v", err)
		return true
	}

	return mms.Value.(bool)
}

// For messages storage with corresponding 'modemId' do:
// - Spawns message handlers.
// - Fills undownloaded map.
func (mediator *Mediator) initializeMessages(modemId string) {
	log.Printf("jezek - Mediator.initializeMessages(%s): start", modemId)
	defer log.Printf("jezek - Mediator.initializeMessages(%s): end", modemId)
	uuids := storage.GetStoredUUIDs()

	log.Printf("Found %d messages in storage", len(uuids))
	for _, uuid := range uuids {
		log.Printf("jezek - checking uuid: %s", uuid)
		mmsState, err := storage.GetMMSState(uuid)
		if err != nil {
			log.Printf("Error checking state of message stored under UUID: %s : %v", uuid, err)
			if err := storage.Destroy(uuid); err != nil {
				log.Printf("Error destroying faulty message: %v", err)
			}
			continue
		}
		log.Printf("jezek - %#v", mmsState)

		if !mmsState.IsIncoming() {
			log.Printf("Message %s is not an incoming message. State: %s", uuid, mmsState.State)
			continue
		}

		// Housekeeping. Delete all old stored incoming messages, which are missing the ModemId.
		if mmsState.ModemId == "" {
			log.Printf("Message %s is an old incoming message with state %s, no need to store, deleting.", uuid, mmsState.State)
			if err := storage.Destroy(uuid); err != nil {
				log.Printf("Error destroying old message: %v", err)
			}
			continue
		}

		if modemId != mmsState.ModemId {
			log.Printf("jezek - message modem id (%s) doesn't match added modem (%s)", mmsState.ModemId, modemId)
			continue
		}
		// Just log any irregularities here.
		if mmsState.MNotificationInd == nil {
			log.Printf("Stored message doesn't contain MNotificationInd")
		}
		if mmsState.MNotificationInd.TransactionId == "" {
			log.Printf("Stored message's MNotificationInd's TransactionId is empty")
		}
		if mediator.telepathyService == nil {
			log.Printf("There is no telepathy service in mediator, will not spawn handlers for this message.")
		}

		checkExpiredAndHandle := func() bool {
			if mmsState.MNotificationInd == nil || !mmsState.MNotificationInd.Expired() {
				return false
			}

			log.Printf("jezek - mNotificationInd is expired")
			// MNotificationInd is expired, destroy in storage & notify telepathy-ofono service.
			if err := storage.Destroy(uuid); err != nil {
				log.Printf("Error destroying expired message: %v", err)
			}
			if mediator.telepathyService != nil {
				//TODO:jezek - Test status change seems not to work on restart. No error, but not working.
				//if err := mediator.telepathyService.MessageHandle(uuid); err != nil {
				//	log.Printf("Error starting message %s handlers to alter expired message", uuid)
				//}
				//if err := mediator.telepathyService.MessageStatusChanged(uuid, telepathy.PERMANENT_ERROR); err != nil {
				//	log.Printf("Error changing message state: %v", err)
				//}
				//if err := mediator.telepathyService.MessageDestroy(uuid); err != nil {
				//	log.Printf("Error destroying message handlers: %v", err)
				//}

				if err := mediator.telepathyService.SingnalMessageRemoved(mediator.telepathyService.GenMessagePath(uuid)); err != nil {
					log.Printf("Error sending signal that message was removed: %v", err)
				}
			}
			return true
		}

		switch mmsState.State {
		case storage.NOTIFICATION:
			// Message download failed, error was probably communicated to telepathy-ofono.
			// It is now up to user to initiate redownload or there is a possibility, that a new notification with the same TransactionId arrives from MMS provider.
			//TODO:jezek - Store more info to be sure if error was communicated and act according.

			if checkExpiredAndHandle() {
				// Message is expired, don't continue.
				break
			}

			// Add to undownloaded, to not communicate possible error to telepathy-ofono again, on possible message notification from mobile provider.
			if mmsState.MNotificationInd != nil && mmsState.MNotificationInd.TransactionId != "" {
				log.Printf("jezek - adding message to undownloaded")
				mediator.undownloaded[mmsState.MNotificationInd.TransactionId] = uuid
			}
			// Spawn interface listener to listen for redownload requests.
			if mediator.telepathyService != nil {
				log.Printf("jezek - spawning handlers for message")
				if err := mediator.telepathyService.MessageHandle(uuid, true); err != nil {
					log.Printf("Error starting message %s handlers of message with state %v", uuid, mmsState.State)
				}
			}
			break

		case storage.DOWNLOADED:
			// Message download was successful, but there was some decoding error, which was probably communicated to telepathy-ofono.
			// It is now up to user to initiate redownload (if he has the choice) or there is a possibility, that a new notification with the same TransactionId arrives from MMS provider.
			//TODO:jezek - Store more info to be sure if error was communicated and act according.

			if checkExpiredAndHandle() {
				// Message is expired, don't continue.
				break
			}

			// Add to undownloaded, to not communicate possible error to telepathy-ofono again, on possible message notification from mobile provider.
			if mmsState.MNotificationInd != nil && mmsState.MNotificationInd.TransactionId != "" {
				log.Printf("jezek - adding message to undownloaded")
				mediator.undownloaded[mmsState.MNotificationInd.TransactionId] = uuid
			}
			// Spawn interface listener to listen for read/delete requests.
			if mediator.telepathyService != nil {
				log.Printf("jezek - spawning handlers for message")
				if err := mediator.telepathyService.MessageHandle(uuid, false); err != nil {
					log.Printf("Error starting message %s handlers of message with state %v", uuid, mmsState.State)
				}
			}
			//TODO:jezek - Try communicate and acknowledge DOWNLOADED message.
			break

		case storage.RECEIVED:
			// Message download was successful, the message was decoded and forwarded to telepathy-ofono, but MMS provider was not notified.
			// There is a possibility, that a new notification with the same TransactionId arrives from MMS provider.
			//TODO:jezek - Try to notify MMS provider and if it fails, add to undownloaded, to not communicate possible error to telepathy-ofono again, on possible message notification from mobile provider.

			if checkExpiredAndHandle() {
				// Message is expired, don't continue.
				break
			}

			// Add to undownloaded, to not communicate possible error to telepathy-ofono again, on possible message notification from mobile provider.
			if mmsState.MNotificationInd != nil && mmsState.MNotificationInd.TransactionId != "" {
				log.Printf("jezek - adding message to undownloaded")
				mediator.undownloaded[mmsState.MNotificationInd.TransactionId] = uuid
			}
			// Spawn interface listener to listen for read/delete requests.
			if mediator.telepathyService != nil {
				log.Printf("jezek - spawning handlers for message")
				if err := mediator.telepathyService.MessageHandle(uuid, false); err != nil {
					log.Printf("Error starting message %s handlers of message with state %v", uuid, mmsState.State)
				}
			}
			break

		case storage.RESPONDED:
			// Message download was successful, the message was decoded and forwarded to telepathy-ofono and MMS provider was notified.
			// Get message from history service and if read or not exist, delete and don't spawn handlers.

			if mediator.telepathyService == nil {
				break
			}

			eventId := string(mediator.telepathyService.GenMessagePath(uuid))
			hsMessage, err := mediator.telepathyService.GetMessage(eventId)
			if err != nil {
				log.Printf("Error getting message %s from HistoryService: %v", eventId, err)
				break
			}
			log.Printf("jezek - hsMessage: %v", hsMessage)

			// If message is doesn't exist, break (don't spawn handlers).
			if len(hsMessage) == 0 {
				log.Printf("Message %s doesn't exist in HistoryService, no need to store, deleting.", uuid)
				if err := storage.Destroy(uuid); err != nil {
					log.Printf("Error destroying message missing in HistoryService: %v", err)
				}
				break
			}

			// If message is marked as read, break (don't spawn handlers).
			if unread, ok := hsMessage["newEvent"].Value.(bool); !ok {
				log.Printf("HistoryService returned a message %s with unusual newEvent field (expecting bool): %#v", eventId, hsMessage["newEvent"])
				break
			} else {
				if unread == false {
					log.Printf("Message %s is marked as read in HistoryService, no need to store, deleting.", uuid)
					if err := storage.Destroy(uuid); err != nil {
						log.Printf("Error destroying message read in HistoryService: %v", err)
					}
					break
				}
			}
			// Spawn interface listener to listen for read/delete requests.
			log.Printf("jezek - spawning handlers for message")
			if err := mediator.telepathyService.MessageHandle(uuid, false); err != nil {
				log.Printf("Error starting message %s handlers of message with state %v", uuid, mmsState.State)
			}
			break

		default:
			log.Printf("Unknown MMSState.State: %s", mmsState.State)
			break
		}

		//TODO:jezek - Telepathy-ofono service should spawn dbus listeners for MarkRead/Delete requests for unread messages os startup.
	}

}
