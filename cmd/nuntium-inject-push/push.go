package main

import (
	"bytes"
	"fmt"

	"github.com/ubports/nuntium/mms"
	"launchpad.net/go-dbus/v1"
)

const (
	pushInterface string = "org.ofono.PushNotificationAgent"
	pushMethod    string = "ReceiveNotification"
)

func getMNotificationIndPayload(args mainFlags) []byte {
	from := mNotificationIndFrom
	if args.SenderNotification != "" {
		from = bytes.Join(
			[][]byte{
				// From + size + token address present +
				[]byte{0x80 + mms.FROM, byte(len(args.SenderNotification)) + 12, mms.TOKEN_ADDRESS_PRESENT},
				// + sender +
				[]byte(args.SenderNotification),
				// + "/TYPE=PLMN|0"
				[]byte{0x2f, 0x54, 0x59, 0x50, 0x45, 0x3d, 0x50, 0x4c, 0x4d, 0x4e, 0x00},
			},
			nil,
		)
	}

	return bytes.Join(
		[][]byte{
			mNotificationIndHeader,
			mNotificationIndVersion,
			from,
			mNotificationIndClass,
			mNotificationIndSize,
			mNotificationIndExpire,
			mNotificationIndContentLocation,
		},
		nil,
	)
}

var mNotificationInd = bytes.Join([][]byte{
	mNotificationIndHeader,
	mNotificationIndVersion,
	mNotificationIndFrom,
	mNotificationIndClass,
	mNotificationIndSize,
	mNotificationIndExpire,
	mNotificationIndContentLocation,
}, nil)

var mNotificationIndHeader = []byte{
	//             &     a     p     p     l     i     c     a     t     i     o
	0x01, 0x06, 0x26, 0x61, 0x70, 0x70, 0x6C, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f,
	// n     /     v     n     d     .     w     a     p     .     m     m     s
	0x6e, 0x2f, 0x76, 0x6e, 0x64, 0x2e, 0x77, 0x61, 0x70, 0x2e, 0x6d, 0x6d, 0x73,
	// -     m     e     s     s     a     g     e
	0x2d, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x00, 0xaf, 0x84, 0xb4, 0x86,
	//                               m     0     4     B     K     k     s     i
	0xc3, 0x95, 0x8c, 0x82, 0x98, 0x6d, 0x30, 0x34, 0x42, 0x4b, 0x6b, 0x73, 0x69,
	// m     0     5     @     m     m     s     .     p     e     r     s     o
	0x6d, 0x30, 0x35, 0x40, 0x6d, 0x6d, 0x73, 0x2e, 0x70, 0x65, 0x72, 0x73, 0x6f,
	// n     a     l     .     c     o     m     .     a     r
	0x6e, 0x61, 0x6c, 0x2e, 0x63, 0x6f, 0x6d, 0x2e, 0x61, 0x72, 0x00,
}

var mNotificationIndVersion = []byte{
	// Version 1.0
	// 0x8d, 0x90,
	0x80 + mms.X_MMS_MMS_VERSION, mms.MMS_MESSAGE_VERSION_1_0,
}

var mNotificationIndFrom = []byte{
	// From + size + token address present + "+543515924906" +
	// 0x89, 25, 0x80, ...
	0x80 + mms.FROM, 0x19, mms.TOKEN_ADDRESS_PRESENT, 0x2b, 0x35, 0x34, 0x33, 0x35, 0x31, 0x35, 0x39, 0x32, 0x34, 0x39, 0x30, 0x36,
	// "/TYPE=PLMN|0"
	0x2f, 0x54, 0x59, 0x50, 0x45, 0x3d, 0x50, 0x4c, 0x4d, 0x4e, 0x00,
}

var mNotificationIndClass = []byte{
	// Class
	// 0x8a, 0x80,
	0x80 + mms.X_MMS_MESSAGE_CLASS, 0x80,
}

var mNotificationIndSize = []byte{
	// Size + num of bytes of size + actual bytes encodin size
	// 0x8e, 2, ...
	0x80 + mms.X_MMS_MESSAGE_SIZE, 0x02, 0x74, 0x00,
}

var mNotificationIndExpire = []byte{
	// Expire + num of bytes encoding token & expire value + token byte + expire value bytes
	// 0x88, 5, ...
	0x80 + mms.X_MMS_EXPIRY, 0x05, 0x81, 0x03, 0x02, 0xa2, 0xff,
}

var mNotificationIndContentLocation = []byte{
	// Content location + "http://localhost:9191/mms|0"
	0x80 + mms.X_MMS_CONTENT_LOCATION,
	// h     t     t     p     :     /     /
	0x68, 0x74, 0x74, 0x70, 0x3a, 0x2f, 0x2f,
	// l     o     c     a     l     h     o     s     t
	0x6c, 0x6f, 0x63, 0x61, 0x6c, 0x68, 0x6f, 0x73, 0x74,
	// :     9     1     9     1     /     m     m     s
	0x3a, 0x39, 0x31, 0x39, 0x31, 0x2f, 0x6d, 0x6d, 0x73, 0x00,
}

func push(args mainFlags) error {
	conn, err := dbus.Connect(dbus.SystemBus)
	if err != nil {
		return err
	}

	obj := conn.Object(args.EndPoint, "/nuntium")

	info := map[string]*dbus.Variant{"LocalSentTime": &dbus.Variant{"2014-02-05T08:29:55-0300"},
		"Sender": &dbus.Variant{args.Sender}}

	reply, err := obj.Call(pushInterface, pushMethod, getMNotificationIndPayload(args), info)
	if err != nil || reply.Type == dbus.TypeError {
		return fmt.Errorf("notification error: %s", err)
	}

	return nil
}
