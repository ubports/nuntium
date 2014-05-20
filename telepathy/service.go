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
	"log"
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
	Payload    ServicePayload
	Properties map[string]dbus.Variant
	conn       *dbus.Connection
	msgChan    chan *dbus.Message
	identity   string
	outMessage chan *OutgoingMessage
}

type Attachment struct {
	Id        string
	MediaType string
	FilePath  string
	Offset    uint64
	Length    uint64
}

type DataPart struct {
	Id          string
	ContentType string
	Filename    string
}

type OutgoingMessage struct {
	Recipients []string
	Smil       string
	DataParts  []DataPart
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
		Payload:    payload,
		Properties: serviceProperties,
		conn:       conn,
		msgChan:    make(chan *dbus.Message),
		outMessage: outgoingChannel,
		identity:   identity,
	}
	go service.watchDBusMethodCalls()
	conn.RegisterObjectPath(payload.Path, service.msgChan)
	return &service
}

func (service *MMSService) watchDBusMethodCalls() {
	var reply *dbus.Message

	for msg := range service.msgChan {
		if msg.Interface != MMS_SERVICE_DBUS_IFACE {
			log.Println("Received unkown method call on", msg.Interface, msg.Member)
			reply = dbus.NewErrorMessage(msg, "org.freedesktop.DBus.Error.UnknownMethod", "Unknown method")
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
		case "GetProperties":
			reply = dbus.NewMethodReturnMessage(msg)
			if err := reply.AppendArgs(service.Properties); err != nil {
				log.Print("Cannot parse payload data from services")
				reply = dbus.NewErrorMessage(msg, "Error.InvalidArguments", "Cannot parse services")
			}
		case "SendMessage":
			reply = dbus.NewMethodReturnMessage(msg)
			var outMessage OutgoingMessage
			if err := reply.AppendArgs(outMessage.Recipients, &outMessage.Smil, outMessage.DataParts); err != nil {
				log.Print("Cannot parse payload data from services")
				reply = dbus.NewErrorMessage(msg, "Error.InvalidArguments", "Cannot parse services")
			} else {
				service.outMessage <- &outMessage
			}
		default:
			log.Println("Received unkown method call on", msg.Interface, msg.Member)
			reply = dbus.NewErrorMessage(msg, "org.freedesktop.DBus.Error.UnknownMethod", "Unknown method")
		}
		if err := service.conn.Send(reply); err != nil {
			log.Println("Could not send reply:", err)
		}
	}
}

//MessageAdded emits a MessageAdded with the path to the added message which
//is taken as a parameter
func (service *MMSService) MessageAdded(mRetConf *mms.MRetrieveConf) error {
	payload, err := service.parseMessage(mRetConf)
	if err != nil {
		return err
	}
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

//TODO randomly creating a uuid until the download manager does this for us
func (service *MMSService) genMessagePath(uuid string) dbus.ObjectPath {
	return dbus.ObjectPath(MMS_DBUS_PATH + "/" + service.identity + "/" + uuid)
}
