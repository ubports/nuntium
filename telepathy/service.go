/*
 * Copyright 2014 Canonical Ltd.
 *
 * Authors:
 * Sergio Schvezov: sergio.schvezov@cannical.com
 *
 * This file is part of telepathy.
 *
 * mms is free software; you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation; version 3.
 *
 * mms is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package telepathy

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/ubports/nuntium/mms"
	"github.com/ubports/nuntium/storage"
	"launchpad.net/go-dbus/v1"
)

//Payload is used to build the dbus messages; this is a workaround as v1 of go-dbus
//tries to encode and decode private fields.
type Payload struct {
	Path       dbus.ObjectPath
	Properties map[string]dbus.Variant
}

type MMSService struct {
	payload              Payload
	Properties           map[string]dbus.Variant
	conn                 *dbus.Connection
	msgChan              chan *dbus.Message
	messageHandlers      map[dbus.ObjectPath]*MessageInterface
	msgDeleteChan        chan dbus.ObjectPath
	msgRedownloadChan    chan dbus.ObjectPath
	identity             string
	outMessage           chan *OutgoingMessage
	mNotificationIndChan chan<- *mms.MNotificationInd
}

type Attachment struct {
	Id        string
	MediaType string
	FilePath  string
	Offset    uint64
	Length    uint64
}

type OutAttachment struct {
	Id          string
	ContentType string
	FilePath    string
}

type OutgoingMessage struct {
	Recipients  []string
	Attachments []OutAttachment
	Reply       *dbus.Message
}

func NewMMSService(conn *dbus.Connection, modemObjPath dbus.ObjectPath, identity string, outgoingChannel chan *OutgoingMessage, useDeliveryReports bool, mNotificationIndChan chan<- *mms.MNotificationInd) *MMSService {
	properties := make(map[string]dbus.Variant)
	properties[identityProperty] = dbus.Variant{identity}
	serviceProperties := make(map[string]dbus.Variant)
	serviceProperties[useDeliveryReportsProperty] = dbus.Variant{useDeliveryReports}
	serviceProperties[modemObjectPathProperty] = dbus.Variant{modemObjPath}
	payload := Payload{
		Path:       dbus.ObjectPath(MMS_DBUS_PATH + "/" + identity),
		Properties: properties,
	}
	service := MMSService{
		payload:              payload,
		Properties:           serviceProperties,
		conn:                 conn,
		msgChan:              make(chan *dbus.Message),
		msgDeleteChan:        make(chan dbus.ObjectPath),
		msgRedownloadChan:    make(chan dbus.ObjectPath),
		messageHandlers:      make(map[dbus.ObjectPath]*MessageInterface),
		outMessage:           outgoingChannel,
		identity:             identity,
		mNotificationIndChan: mNotificationIndChan,
	}
	go service.watchDBusMethodCalls()
	go service.watchMessageDeleteCalls()
	go service.watchMessageRedownloadCalls()
	conn.RegisterObjectPath(payload.Path, service.msgChan)
	return &service
}

func (*MMSService) getMMSState(objectPath dbus.ObjectPath) (storage.MMSState, error) {
	uuid, err := getUUIDFromObjectPath(objectPath)
	if err != nil {
		return storage.MMSState{}, err
	}

	return storage.GetMMSState(uuid)
}

func (service *MMSService) watchMessageDeleteCalls() {
	for msgObjectPath := range service.msgDeleteChan {
		log.Printf("jezek - MMSService.watchMessageDeleteCalls: msgObjectPath: %v", msgObjectPath)

		if mmsState, err := service.getMMSState(msgObjectPath); err == nil {
			if mmsState.State == storage.NOTIFICATION {
				log.Printf("jezek - MMSService.watchMessageDeleteCalls: Message not fully downloaded, not deleting.")
				continue
			}
		} else {
			log.Printf("jezek - MMSService.watchMessageRedownloadCalls: error retrieving message state: %v", err)
		}

		if err := service.MessageRemoved(msgObjectPath); err != nil {
			log.Print("Failed to delete ", msgObjectPath, ": ", err)
		}
	}
}

func (service *MMSService) watchMessageRedownloadCalls() {
	for msgObjectPath := range service.msgRedownloadChan {
		log.Printf("jezek - MMSService.watchMessageRedownloadCalls: msgObjectPath: %v", msgObjectPath)

		mmsState, err := service.getMMSState(msgObjectPath)
		if err != nil {
			log.Printf("jezek - MMSService.watchMessageRedownloadCalls: error retrieving message state: %v", err)
			continue
		}
		if mmsState.State != storage.NOTIFICATION {
			log.Printf("jezek - MMSService.watchMessageRedownloadCalls: message was already downloaded")
			continue
		}
		if mmsState.MNotificationInd == nil {
			log.Printf("jezek - MMSService.watchMessageRedownloadCalls: no mNotificationInd found.")
			continue
		}
		log.Printf("jezek - MMSService.watchMessageRedownloadCalls: mNotificationInd: %#v", mmsState.MNotificationInd)

		if err := service.MessageRemoved(msgObjectPath); err != nil {
			//TODO:jezek - just log?, can some panic ocure after this?
			log.Print("Failed to delete ", msgObjectPath, ": ", err)
		}
		newMNotificationInd := mmsState.MNotificationInd
		newMNotificationInd.RedownloadOfUUID = mmsState.MNotificationInd.UUID
		newMNotificationInd.UUID = mms.GenUUID()
		storage.Create(mmsState.ModemId, newMNotificationInd)
		log.Printf("jezek - MMSService.watchMessageRedownloadCalls: new mNotificationInd new: %#v", newMNotificationInd)
		service.mNotificationIndChan <- newMNotificationInd
	}
}

func (service *MMSService) watchDBusMethodCalls() {
	log.Printf("jezek - service %v: watchDBusMethodCalls(): start", service.identity)
	defer log.Printf("jezek - service %v: watchDBusMethodCalls(): end", service.identity)
	for msg := range service.msgChan {
		log.Printf("jezek - service %v: watchDBusMethodCalls(): Received message: %s - %v", service.identity, msg.Member, msg)
		var reply *dbus.Message
		if msg.Interface != MMS_SERVICE_DBUS_IFACE {
			log.Println("Received unkown method call on", msg.Interface, msg.Member)
			reply = dbus.NewErrorMessage(
				msg,
				"org.freedesktop.DBus.Error.UnknownInterface",
				fmt.Sprintf("No such interface '%s' at object path '%s'", msg.Interface, msg.Path))
			//TODO:jezek Send the reply?
			continue
		}
		switch msg.Member {
		case "GetMessages":
			reply = dbus.NewMethodReturnMessage(msg)
			//TODO implement store and forward
			var payload []Payload
			if err := reply.AppendArgs(payload); err != nil {
				log.Print("Cannot parse payload data from services")
				reply = dbus.NewErrorMessage(msg, "Error.InvalidArguments", "Cannot parse services")
			}
			if err := service.conn.Send(reply); err != nil {
				log.Println("Could not send reply:", err)
			}
		case "GetProperties":
			reply = dbus.NewMethodReturnMessage(msg)
			if pc, err := service.GetPreferredContext(); err == nil {
				service.Properties[preferredContextProperty] = dbus.Variant{pc}
			} else {
				// Using "/" as an invalid 'path' even though it could be considered 'incorrect'
				service.Properties[preferredContextProperty] = dbus.Variant{dbus.ObjectPath("/")}
			}
			if err := reply.AppendArgs(service.Properties); err != nil {
				log.Print("Cannot parse payload data from services")
				reply = dbus.NewErrorMessage(msg, "Error.InvalidArguments", "Cannot parse services")
			}
			if err := service.conn.Send(reply); err != nil {
				log.Println("Could not send reply:", err)
			}
		case "SetProperty":
			if err := service.setProperty(msg); err != nil {
				log.Println("Property set failed:", err)
				reply = dbus.NewErrorMessage(msg, "Error.InvalidArguments", err.Error())
			} else {
				reply = dbus.NewMethodReturnMessage(msg)
			}
			if err := service.conn.Send(reply); err != nil {
				log.Println("Could not send reply:", err)
			}
		case "SendMessage":
			var outMessage OutgoingMessage
			outMessage.Reply = dbus.NewMethodReturnMessage(msg)
			if err := msg.Args(&outMessage.Recipients, &outMessage.Attachments); err != nil {
				log.Print("Cannot parse payload data from services")
				reply = dbus.NewErrorMessage(msg, "Error.InvalidArguments", "Cannot parse New Message")
				if err := service.conn.Send(reply); err != nil {
					log.Println("Could not send reply:", err)
				}
			} else {
				service.outMessage <- &outMessage
			}
		default:
			log.Println("Received unkown method call on", msg.Interface, msg.Member)
			reply = dbus.NewErrorMessage(
				msg,
				"org.freedesktop.DBus.Error.UnknownMethod",
				fmt.Sprintf("No such method '%s' at object path '%s'", msg.Member, msg.Path))
			if err := service.conn.Send(reply); err != nil {
				log.Println("Could not send reply:", err)
			}
		}
	}
}

func getUUIDFromObjectPath(objectPath dbus.ObjectPath) (string, error) {
	str := string(objectPath)
	defaultError := fmt.Errorf("%s is not a proper object path for a Message", str)
	if str == "" {
		return "", defaultError
	}
	uuid := filepath.Base(str)
	if uuid == "" || uuid == ".." || uuid == "." {
		return "", defaultError
	}
	return uuid, nil
}

func (service *MMSService) SetPreferredContext(context dbus.ObjectPath) error {
	// make set a noop if we are setting the same thing
	if pc, err := service.GetPreferredContext(); err != nil && context == pc {
		return nil
	}

	if err := storage.SetPreferredContext(service.identity, context); err != nil {
		return err
	}
	signal := dbus.NewSignalMessage(service.payload.Path, MMS_SERVICE_DBUS_IFACE, propertyChangedSignal)
	if err := signal.AppendArgs(preferredContextProperty, dbus.Variant{context}); err != nil {
		return err
	}
	return service.conn.Send(signal)
}

func (service *MMSService) GetPreferredContext() (dbus.ObjectPath, error) {
	return storage.GetPreferredContext(service.identity)
}

func (service *MMSService) setProperty(msg *dbus.Message) error {
	var propertyName string
	var propertyValue dbus.Variant
	if err := msg.Args(&propertyName, &propertyValue); err != nil {
		return err
	}

	switch propertyName {
	case preferredContextProperty:
		preferredContextObjectPath := dbus.ObjectPath(reflect.ValueOf(propertyValue.Value).String())
		service.Properties[preferredContextProperty] = dbus.Variant{preferredContextObjectPath}
		return service.SetPreferredContext(preferredContextObjectPath)
	default:
		errors.New("property cannot be set")
	}
	return errors.New("unhandled property")
}

//MessageRemoved emits the MessageRemoved signal with the path of the removed
//message.
//It also actually removes the message from storage.
func (service *MMSService) MessageRemoved(objectPath dbus.ObjectPath) error {
	service.messageHandlers[objectPath].Close()
	delete(service.messageHandlers, objectPath)

	uuid, err := getUUIDFromObjectPath(objectPath)
	if err != nil {
		return err
	}
	if err := storage.Destroy(uuid); err != nil {
		return err
	}

	signal := dbus.NewSignalMessage(service.payload.Path, MMS_SERVICE_DBUS_IFACE, messageRemovedSignal)
	if err := signal.AppendArgs(objectPath); err != nil {
		return err
	}
	if err := service.conn.Send(signal); err != nil {
		return err
	}
	return nil
}

func (service *MMSService) IncomingMessageFailAdded(mNotificationInd *mms.MNotificationInd) error {
	//just handle that mms as an empty MMS
	params := make(map[string]dbus.Variant)

	// Signal path:
	// https://github.com/ubports/telepathy-ofono/blob/040321101e7bfe5950934a1b718875f3fe29c495/mmsdservice.cpp#L118
	// https://github.com/ubports/telepathy-ofono/blob/040321101e7bfe5950934a1b718875f3fe29c495/connection.cpp#L518
	// https://github.com/ubports/telepathy-ofono/blob/040321101e7bfe5950934a1b718875f3fe29c495/connection.cpp#L423
	// https://github.com/ubports/telepathy-ofono/blob/db5e35b68f244d007468b8de2d9ad9998a2c8bd7/ofonotextchannel.cpp#L473
	// https://github.com/TelepathyIM/telepathy-qt/blob/7cf3e35fdf6cf7ea7d8fc301eae04fe43930b17f/TelepathyQt/base-channel.cpp#L460
	// https://github.com/ubports/history-service/blob/8285a4a3174b84a04f00d600fff99905aec6c4e2/daemon/historydaemon.cpp#L1023
	params["Status"] = dbus.Variant{"received"}
	date := time.Now().Format(time.RFC3339)
	params["Date"] = dbus.Variant{date}
	params["Error"] = dbus.Variant{"Error"}

	sender := mNotificationInd.From
	if strings.HasSuffix(mNotificationInd.From, PLMN) {
		params["Sender"] = dbus.Variant{sender[:len(sender)-len(PLMN)]}
	}

	payload := Payload{Path: service.GenMessagePath(mNotificationInd.UUID), Properties: params}

	if mNotificationInd.RedownloadOfUUID != "" {
		payload.Properties["DeleteEvent"] = dbus.Variant{string(service.GenMessagePath(mNotificationInd.RedownloadOfUUID))}
	}

	service.messageHandlers[payload.Path] = NewMessageInterface(service.conn, payload.Path, service.msgDeleteChan, service.msgRedownloadChan)
	return service.MessageAdded(&payload)
}

//IncomingMessageAdded emits a MessageAdded with the path to the added message which
//is taken as a parameter and creates an object path on the message interface.
func (service *MMSService) IncomingMessageAdded(mRetConf *mms.MRetrieveConf, mNotificationInd *mms.MNotificationInd) error {
	payload, err := service.parseMessage(mRetConf)
	if err != nil {
		return err
	}

	if mNotificationInd.RedownloadOfUUID != "" {
		payload.Properties["DeleteEvent"] = dbus.Variant{string(service.GenMessagePath(mNotificationInd.RedownloadOfUUID))}
	}

	service.messageHandlers[payload.Path] = NewMessageInterface(service.conn, payload.Path, service.msgDeleteChan, service.msgRedownloadChan)
	return service.MessageAdded(&payload)
}

//MessageAdded emits a MessageAdded with the path to the added message which
//is taken as a parameter
func (service *MMSService) MessageAdded(msgPayload *Payload) error {
	log.Printf("jezek - service %v: MessageAdded(): payload: %v", service.identity, msgPayload)
	signal := dbus.NewSignalMessage(service.payload.Path, MMS_SERVICE_DBUS_IFACE, messageAddedSignal)
	if err := signal.AppendArgs(msgPayload.Path, msgPayload.Properties); err != nil {
		return err
	}
	return service.conn.Send(signal)
}

func (service *MMSService) isService(identity string) bool {
	path := dbus.ObjectPath(MMS_DBUS_PATH + "/" + identity)
	if path == service.payload.Path {
		return true
	}
	return false
}

func (service *MMSService) Close() {
	service.conn.UnregisterObjectPath(service.payload.Path)
	close(service.msgChan)
	close(service.msgDeleteChan)
	close(service.msgRedownloadChan)
}

func (service *MMSService) parseMessage(mRetConf *mms.MRetrieveConf) (Payload, error) {
	params := make(map[string]dbus.Variant)
	params["Status"] = dbus.Variant{"received"}
	//TODO retrieve date correctly
	date := parseDate(mRetConf.Date)
	params["Date"] = dbus.Variant{date}
	if mRetConf.Subject != "" {
		params["Subject"] = dbus.Variant{mRetConf.Subject}
	}
	sender := mRetConf.From
	if strings.HasSuffix(mRetConf.From, PLMN) {
		params["Sender"] = dbus.Variant{sender[:len(sender)-len(PLMN)]}
	}

	params["Recipients"] = dbus.Variant{parseRecipients(strings.Join(mRetConf.To, ","))}
	if smil, err := mRetConf.GetSmil(); err == nil {
		params["Smil"] = dbus.Variant{smil}
	}
	var attachments []Attachment
	dataParts := mRetConf.GetDataParts()
	for i := range dataParts {
		var filePath string
		if f, err := storage.GetMMS(mRetConf.UUID); err == nil {
			filePath = f
		} else {
			return Payload{}, err
		}
		attachment := Attachment{
			Id:        dataParts[i].ContentId,
			MediaType: dataParts[i].MediaType,
			FilePath:  filePath,
			Offset:    uint64(dataParts[i].Offset),
			Length:    uint64(len(dataParts[i].Data)),
		}
		attachments = append(attachments, attachment)
	}
	params["Attachments"] = dbus.Variant{attachments}
	payload := Payload{Path: service.GenMessagePath(mRetConf.UUID), Properties: params}
	return payload, nil
}

func parseDate(unixTime uint64) string {
	const layout = "2014-03-30T18:15:30-0300"
	date := time.Unix(int64(unixTime), 0)
	return date.Format(time.RFC3339)
}

func parseRecipients(to string) []string {
	recipients := strings.Split(to, ",")
	for i := range recipients {
		if strings.HasSuffix(recipients[i], PLMN) {
			recipients[i] = recipients[i][:len(recipients[i])-len(PLMN)]
		}
	}
	return recipients
}

func (service *MMSService) MessageDestroy(uuid string) error {
	msgObjectPath := service.GenMessagePath(uuid)
	if msgInterface, ok := service.messageHandlers[msgObjectPath]; ok {
		msgInterface.Close()
		delete(service.messageHandlers, msgObjectPath)
	}
	return fmt.Errorf("no message interface handler for object path %s", msgObjectPath)
}

func (service *MMSService) MessageStatusChanged(uuid, status string) error {
	msgObjectPath := service.GenMessagePath(uuid)
	if msgInterface, ok := service.messageHandlers[msgObjectPath]; ok {
		return msgInterface.StatusChanged(status)
	}
	return fmt.Errorf("no message interface handler for object path %s", msgObjectPath)
}

func (service *MMSService) ReplySendMessage(reply *dbus.Message, uuid string) (dbus.ObjectPath, error) {
	msgObjectPath := service.GenMessagePath(uuid)
	reply.AppendArgs(msgObjectPath)
	if err := service.conn.Send(reply); err != nil {
		return "", err
	}
	msg := NewMessageInterface(service.conn, msgObjectPath, service.msgDeleteChan, service.msgRedownloadChan)
	service.messageHandlers[msgObjectPath] = msg
	service.MessageAdded(msg.GetPayload())
	return msgObjectPath, nil
}

//TODO randomly creating a uuid until the download manager does this for us
func (service *MMSService) GenMessagePath(uuid string) dbus.ObjectPath {
	return dbus.ObjectPath(MMS_DBUS_PATH + "/" + service.identity + "/" + uuid)
}

// Creates handlers for message.
// If already handled, prints log and returns.
func (service *MMSService) HandleMessage(uuid string) {
	path := service.GenMessagePath(uuid)
	if _, ok := service.messageHandlers[path]; ok {
		log.Printf("Message %s already handled", uuid)
		return
	}
	service.messageHandlers[path] = NewMessageInterface(service.conn, path, service.msgDeleteChan, service.msgRedownloadChan)
}
