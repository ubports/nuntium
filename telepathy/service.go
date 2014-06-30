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
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	"launchpad.net/go-dbus/v1"
	"launchpad.net/nuntium/mms"
	"launchpad.net/nuntium/storage"
)

//ServicePayload is used to build the dbus messages; this is a workaround as v1 of go-dbus
//tries to encode and decode private fields.
type ServicePayload struct {
	Path       dbus.ObjectPath
	Properties map[string]dbus.Variant
}

type MMSService struct {
	Payload         ServicePayload
	Properties      map[string]dbus.Variant
	conn            *dbus.Connection
	msgChan         chan *dbus.Message
	messageHandlers map[dbus.ObjectPath]*MessageInterface
	msgDeleteChan   chan dbus.ObjectPath
	identity        string
	outMessage      chan *OutgoingMessage
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

func NewMMSService(conn *dbus.Connection, identity string, outgoingChannel chan *OutgoingMessage, useDeliveryReports bool) *MMSService {
	properties := make(map[string]dbus.Variant)
	properties[IDENTITY] = dbus.Variant{identity}
	serviceProperties := make(map[string]dbus.Variant)
	serviceProperties[USE_DELIVERY_REPORTS] = dbus.Variant{useDeliveryReports}
	payload := ServicePayload{
		Path:       dbus.ObjectPath(MMS_DBUS_PATH + "/" + identity),
		Properties: properties,
	}
	service := MMSService{
		Payload:         payload,
		Properties:      serviceProperties,
		conn:            conn,
		msgChan:         make(chan *dbus.Message),
		msgDeleteChan:   make(chan dbus.ObjectPath),
		messageHandlers: make(map[dbus.ObjectPath]*MessageInterface),
		outMessage:      outgoingChannel,
		identity:        identity,
	}
	go service.watchDBusMethodCalls()
	go service.watchMessageDeleteCalls()
	conn.RegisterObjectPath(payload.Path, service.msgChan)
	return &service
}

func (service *MMSService) watchMessageDeleteCalls() {
	for msgObjectPath := range service.msgDeleteChan {
		if err := service.MessageRemoved(msgObjectPath); err != nil {
			log.Print("Failed to delete ", msgObjectPath, ": ", err)
		}
	}
}

func (service *MMSService) watchDBusMethodCalls() {
	for msg := range service.msgChan {
		var reply *dbus.Message
		if msg.Interface != MMS_SERVICE_DBUS_IFACE {
			log.Println("Received unkown method call on", msg.Interface, msg.Member)
			reply = dbus.NewErrorMessage(
				msg,
				"org.freedesktop.DBus.Error.UnknownInterface",
				fmt.Sprintf("No such interface '%s' at object path '%s'", msg.Interface, msg.Path))
			continue
		}
		switch msg.Member {
		case "GetMessages":
			reply = dbus.NewMethodReturnMessage(msg)
			//TODO implement store and forward
			var payload []ServicePayload
			if err := reply.AppendArgs(payload); err != nil {
				log.Print("Cannot parse payload data from services")
				reply = dbus.NewErrorMessage(msg, "Error.InvalidArguments", "Cannot parse services")
			}
			if err := service.conn.Send(reply); err != nil {
				log.Println("Could not send reply:", err)
			}
		case "GetProperties":
			reply = dbus.NewMethodReturnMessage(msg)
			if err := reply.AppendArgs(service.Properties); err != nil {
				log.Print("Cannot parse payload data from services")
				reply = dbus.NewErrorMessage(msg, "Error.InvalidArguments", "Cannot parse services")
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

	signal := dbus.NewSignalMessage(service.Payload.Path, MMS_SERVICE_DBUS_IFACE, MESSAGE_REMOVED)
	if err := signal.AppendArgs(objectPath); err != nil {
		return err
	}
	if err := service.conn.Send(signal); err != nil {
		return err
	}
	return nil
}

//MessageAdded emits a MessageAdded with the path to the added message which
//is taken as a parameter
func (service *MMSService) MessageAdded(mRetConf *mms.MRetrieveConf) error {
	payload, err := service.parseMessage(mRetConf)
	if err != nil {
		return err
	}
	service.messageHandlers[payload.Path] = NewMessageInterface(service.conn, payload.Path, service.msgDeleteChan)
	signal := dbus.NewSignalMessage(service.Payload.Path, MMS_SERVICE_DBUS_IFACE, MESSAGE_ADDED)
	if err := signal.AppendArgs(payload.Path, payload.Properties); err != nil {
		return err
	}
	if err := service.conn.Send(signal); err != nil {
		return err
	}
	return nil
}

func (service *MMSService) isService(identity string) bool {
	path := dbus.ObjectPath(MMS_DBUS_PATH + "/" + identity)
	if path == service.Payload.Path {
		return true
	}
	return false
}

func (service *MMSService) Close() {
	service.conn.UnregisterObjectPath(service.Payload.Path)
	close(service.msgChan)
	close(service.msgDeleteChan)
}

func (service *MMSService) parseMessage(mRetConf *mms.MRetrieveConf) (ServicePayload, error) {
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

	params["Recipients"] = dbus.Variant{parseRecipients(mRetConf.To)}
	if smil, err := mRetConf.GetSmil(); err == nil {
		params["Smil"] = dbus.Variant{smil}
	} else {
		return ServicePayload{}, err
	}
	var attachments []Attachment
	dataParts := mRetConf.GetDataParts()
	for i := range dataParts {
		var filePath string
		if f, err := storage.GetMMS(mRetConf.UUID); err == nil {
			filePath = f
		} else {
			return ServicePayload{}, err
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
	payload := ServicePayload{Path: service.genMessagePath(mRetConf.UUID), Properties: params}
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

func (service *MMSService) MessageStatusChanged(uuid, status string) error {
	msgObjectPath := service.genMessagePath(uuid)
	if msgInterface, ok := service.messageHandlers[msgObjectPath]; ok {
		return msgInterface.StatusChanged(status)
	}
	return fmt.Errorf("no message interface handler for object path %s", msgObjectPath)
}

func (service *MMSService) ReplySendMessage(reply *dbus.Message, uuid string) (dbus.ObjectPath, error) {
	msgObjectPath := service.genMessagePath(uuid)
	reply.AppendArgs(msgObjectPath)
	if err := service.conn.Send(reply); err != nil {
		return "", err
	}
	service.messageHandlers[msgObjectPath] = NewMessageInterface(service.conn, msgObjectPath, service.msgDeleteChan)
	return msgObjectPath, nil
}

//TODO randomly creating a uuid until the download manager does this for us
func (service *MMSService) genMessagePath(uuid string) dbus.ObjectPath {
	return dbus.ObjectPath(MMS_DBUS_PATH + "/" + service.identity + "/" + uuid)
}
