package main

import (
	"bytes"

	"github.com/ubports/nuntium/mms"
)

var mRetrieveConf = bytes.Join([][]byte{
	mRetrieveConfType,
	mRetrieveConfTransactionId,
	mRetrieveConfMMSVersion,
	mRetrieveConfMessageId,
	mRetrieveConfDate,
	mRetrieveConfFrom,
	mRetrieveConfContentType,
	mRetrieveConfAttachments,
}, nil)

var mRetrieveConfType = []byte{
	// Type m-Retrieve.conf
	// 0x8c, 0x82
	0x80 + mms.X_MMS_MESSAGE_TYPE, mms.TYPE_RETRIEVE_CONF,
}
var mRetrieveConfTransactionId = []byte{
	// Transaction Id "123456789\0"
	// 0x98, ...
	0x80 + mms.X_MMS_TRANSACTION_ID, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x00,
}
var mRetrieveConfMMSVersion = []byte{
	// Version 1.3
	// 0x8d, 0x93
	0x80 + mms.X_MMS_MMS_VERSION, mms.MMS_MESSAGE_VERSION_1_3,
}
var mRetrieveConfMessageId = []byte{
	// Message Id "abdefghij\0"
	0x80 + mms.MESSAGE_ID, 0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x70, 0x71, 0x72, 0x72, 0x73, 0x74, 0x00,
}
var mRetrieveConfDate = []byte{
	// Date
	0x80 + mms.DATE, 0x04, 0x54, 0x5a, 0xc0, 0x37,
}
var mRetrieveConfFrom = []byte{
	// From + size + token address present + "01189998819991197253" +
	// 0x89, 32, 0x80, ...
	0x80 + mms.FROM, 0x20, mms.TOKEN_ADDRESS_PRESENT, 0x30, 0x31, 0x31, 0x38, 0x39, 0x39, 0x39, 0x38, 0x38, 0x31, 0x39, 0x39, 0x39, 0x31, 0x31, 0x39, 0x37, 0x32, 0x35, 0x33,
	// + "/TYPE=PLMN\0"
	0x2f, 0x54, 0x59, 0x50, 0x45, 0x3d, 0x50, 0x4c, 0x4d, 0x4e, 0x00,
}
var mRetrieveConfContentType = []byte{
	// Content Type application/vnd.wap.multipart.related
	0x80 + mms.CONTENT_TYPE, 0x1b,
}

var mRetrieveConfAttachments = []byte{
	// Attachments
	0xb3,
	0x89, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x73, 0x6d, 0x69, 0x6c,
	0x00, 0x8a, 0x3c, 0x73, 0x6d, 0x69, 0x6c, 0x3e, 0x00, 0x03, 0x2f, 0x83, 0x2b, 0x1b, 0x61, 0x70, 0x70,
	0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x73, 0x6d, 0x69, 0x6c, 0x00, 0x85, 0x73, 0x6d,
	0x69, 0x6c, 0x2e, 0x78, 0x6d, 0x6c, 0x00, 0x8e, 0x73, 0x6d, 0x69, 0x6c, 0x2e, 0x78, 0x6d, 0x6c, 0x00,
	0xc0, 0x22, 0x3c, 0x73, 0x6d, 0x69, 0x6c, 0x3e, 0x00, 0x3c, 0x73, 0x6d, 0x69, 0x6c, 0x3e, 0x20, 0x20,
	0x20, 0x3c, 0x68, 0x65, 0x61, 0x64, 0x3e, 0x20, 0x20, 0x20, 0x20, 0x20, 0x3c, 0x6c, 0x61, 0x79, 0x6f,
	0x75, 0x74, 0x3e, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x3c, 0x72, 0x65, 0x67, 0x69,
	0x6f, 0x6e, 0x20, 0x69, 0x64, 0x3d, 0x22, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x22, 0x20, 0x77, 0x69, 0x64,
	0x74, 0x68, 0x3d, 0x22, 0x31, 0x30, 0x30, 0x25, 0x22, 0x20, 0x68, 0x65, 0x69, 0x67, 0x68, 0x74, 0x3d,
	0x22, 0x31, 0x30, 0x30, 0x25, 0x22, 0x20, 0x66, 0x69, 0x74, 0x3d, 0x22, 0x6d, 0x65, 0x65, 0x74, 0x22,
	0x20, 0x2f, 0x3e, 0x3c, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x20, 0x69, 0x64, 0x3d, 0x22, 0x54, 0x65,
	0x78, 0x74, 0x22, 0x20, 0x77, 0x69, 0x64, 0x74, 0x68, 0x3d, 0x22, 0x31, 0x30, 0x30, 0x25, 0x22, 0x20,
	0x68, 0x65, 0x69, 0x67, 0x68, 0x74, 0x3d, 0x22, 0x31, 0x30, 0x30, 0x25, 0x22, 0x20, 0x66, 0x69, 0x74,
	0x3d, 0x22, 0x73, 0x63, 0x72, 0x6f, 0x6c, 0x6c, 0x22, 0x20, 0x2f, 0x3e, 0x3c, 0x72, 0x65, 0x67, 0x69,
	0x6f, 0x6e, 0x20, 0x69, 0x64, 0x3d, 0x22, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x22, 0x20, 0x77, 0x69, 0x64,
	0x74, 0x68, 0x3d, 0x22, 0x31, 0x30, 0x30, 0x25, 0x22, 0x20, 0x68, 0x65, 0x69, 0x67, 0x68, 0x74, 0x3d,
	0x22, 0x31, 0x30, 0x30, 0x25, 0x22, 0x20, 0x66, 0x69, 0x74, 0x3d, 0x22, 0x6d, 0x65, 0x65, 0x74, 0x22,
	0x20, 0x2f, 0x3e, 0x20, 0x20, 0x20, 0x20, 0x20, 0x3c, 0x2f, 0x6c, 0x61, 0x79, 0x6f, 0x75, 0x74, 0x3e,
	0x20, 0x20, 0x20, 0x3c, 0x2f, 0x68, 0x65, 0x61, 0x64, 0x3e, 0x20, 0x20, 0x20, 0x3c, 0x62, 0x6f, 0x64,
	0x79, 0x3e, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x3c, 0x70, 0x61, 0x72, 0x20, 0x64, 0x75, 0x72,
	0x3d, 0x22, 0x35, 0x30, 0x30, 0x30, 0x6d, 0x73, 0x22, 0x3e, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20,
	0x3c, 0x69, 0x6d, 0x67, 0x20, 0x73, 0x72, 0x63, 0x3d, 0x22, 0x63, 0x69, 0x64, 0x3a, 0x75, 0x62, 0x75,
	0x6e, 0x74, 0x75, 0x2e, 0x6a, 0x70, 0x67, 0x22, 0x20, 0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x3d, 0x22,
	0x49, 0x6d, 0x61, 0x67, 0x65, 0x22, 0x20, 0x2f, 0x3e, 0x20, 0x20, 0x20, 0x20, 0x20, 0x3c, 0x2f, 0x70,
	0x61, 0x72, 0x3e, 0x3c, 0x70, 0x61, 0x72, 0x20, 0x64, 0x75, 0x72, 0x3d, 0x22, 0x33, 0x73, 0x22, 0x3e,
	0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x20, 0x3c, 0x74, 0x65, 0x78, 0x74, 0x20, 0x73, 0x72, 0x63, 0x3d,
	0x22, 0x63, 0x69, 0x64, 0x3a, 0x74, 0x65, 0x78, 0x74, 0x5f, 0x30, 0x2e, 0x74, 0x78, 0x74, 0x22, 0x20,
	0x72, 0x65, 0x67, 0x69, 0x6f, 0x6e, 0x3d, 0x22, 0x54, 0x65, 0x78, 0x74, 0x22, 0x20, 0x2f, 0x3e, 0x20,
	0x20, 0x20, 0x20, 0x20, 0x3c, 0x2f, 0x70, 0x61, 0x72, 0x3e, 0x20, 0x20, 0x20, 0x3c, 0x2f, 0x62, 0x6f,
	0x64, 0x79, 0x3e, 0x20, 0x3c, 0x2f, 0x73, 0x6d, 0x69, 0x6c, 0x3e, 0x27, 0x95, 0x3a, 0x0d, 0x9e, 0x85,
	0x75, 0x62, 0x75, 0x6e, 0x74, 0x75, 0x2e, 0x6a, 0x70, 0x67, 0x00, 0x8e, 0x75, 0x62, 0x75, 0x6e, 0x74,
	0x75, 0x2e, 0x6a, 0x70, 0x67, 0x00, 0xc0, 0x22, 0x75, 0x62, 0x75, 0x6e, 0x74, 0x75, 0x2e, 0x6a, 0x70,
	0x67, 0x00, 0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 0x4a, 0x46, 0x49, 0x46, 0x00, 0x01, 0x01, 0x02, 0x00,
	0x1c, 0x00, 0x1c, 0x00, 0x00, 0xff, 0xdb, 0x00, 0x43, 0x00, 0x03, 0x02, 0x02, 0x02, 0x02, 0x02, 0x03,
	0x02, 0x02, 0x02, 0x03, 0x03, 0x03, 0x03, 0x04, 0x06, 0x04, 0x04, 0x04, 0x04, 0x04, 0x08, 0x06, 0x06,
	0x05, 0x06, 0x09, 0x08, 0x0a, 0x0a, 0x09, 0x08, 0x09, 0x09, 0x0a, 0x0c, 0x0f, 0x0c, 0x0a, 0x0b, 0x0e,
	0x0b, 0x09, 0x09, 0x0d, 0x11, 0x0d, 0x0e, 0x0f, 0x10, 0x10, 0x11, 0x10, 0x0a, 0x0c, 0x12, 0x13, 0x12,
	0x10, 0x13, 0x0f, 0x10, 0x10, 0x10, 0xff, 0xdb, 0x00, 0x43, 0x01, 0x03, 0x03, 0x03, 0x04, 0x03, 0x04,
	0x08, 0x04, 0x04, 0x08, 0x10, 0x0b, 0x09, 0x0b, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10,
	0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10,
	0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10,
	0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0xff, 0xc0, 0x00, 0x11, 0x08, 0x00, 0x40, 0x00, 0x40, 0x03,
	0x01, 0x22, 0x00, 0x02, 0x11, 0x01, 0x03, 0x11, 0x01, 0xff, 0xc4, 0x00, 0x1b, 0x00, 0x01, 0x01, 0x01,
	0x01, 0x00, 0x03, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x08, 0x07, 0x06, 0x09,
	0x00, 0x02, 0x04, 0x05, 0xff, 0xc4, 0x00, 0x32, 0x10, 0x00, 0x01, 0x03, 0x03, 0x03, 0x03, 0x03, 0x02,
	0x05, 0x04, 0x03, 0x01, 0x00, 0x00, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x11, 0x00, 0x07,
	0x12, 0x08, 0x13, 0x21, 0x31, 0x41, 0x51, 0x62, 0x71, 0x14, 0x22, 0x61, 0x72, 0x91, 0x33, 0x42, 0x82,
	0xa1, 0x15, 0x17, 0x23, 0x32, 0xff, 0xc4, 0x00, 0x19, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x06, 0x07, 0x05, 0x01, 0x02, 0x04, 0xff, 0xc4,
	0x00, 0x2b, 0x11, 0x00, 0x01, 0x02, 0x04, 0x05, 0x03, 0x05, 0x00, 0x03, 0x01, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x01, 0x02, 0x11, 0x03, 0x05, 0x06, 0x21, 0x00, 0x04, 0x12, 0x31, 0x51, 0x13, 0x41, 0x81,
	0x22, 0x32, 0x61, 0x71, 0x91, 0x23, 0xa1, 0xb1, 0x52, 0xff, 0xda, 0x00, 0x0c, 0x03, 0x01, 0x00, 0x02,
	0x11, 0x03, 0x11, 0x00, 0x3f, 0x00, 0xe5, 0x56, 0xb6, 0x5b, 0x73, 0xb5, 0x57, 0x5e, 0xe5, 0xcf, 0x31,
	0xe8, 0x51, 0x52, 0xdc, 0x46, 0x54, 0x04, 0x99, 0xaf, 0x65, 0x2c, 0xb1, 0x9f, 0x93, 0xea, 0xa5, 0x63,
	0xd1, 0x29, 0xc9, 0xfb, 0x0f, 0x3a, 0xf3, 0x6a, 0xb6, 0xe2, 0xa1, 0xb9, 0x97, 0x5b, 0x34, 0x38, 0xcb,
	0x2c, 0x45, 0x6c, 0x77, 0xe6, 0xc9, 0xc6, 0x7b, 0x2c, 0x02, 0x01, 0x23, 0xea, 0x39, 0xc2, 0x47, 0xc9,
	0xf8, 0x07, 0x4e, 0x6a, 0x0d, 0x06, 0x91, 0x6c, 0x52, 0x23, 0x50, 0xa8, 0x50, 0x51, 0x12, 0x0c, 0x44,
	0xf0, 0x69, 0xa4, 0x7b, 0x7c, 0x92, 0x7d, 0x54, 0xa2, 0x7c, 0x95, 0x1f, 0x24, 0xe8, 0xad, 0x47, 0x51,
	0x89, 0x40, 0x10, 0x60, 0x5e, 0x29, 0xfc, 0x48, 0xe4, 0xfc, 0xf0, 0x3c, 0x9f, 0x9a, 0x15, 0x13, 0x44,
	0xaa, 0xa3, 0x51, 0xcd, 0x66, 0xc9, 0x4c, 0x04, 0x96, 0xb6, 0xea, 0x3c, 0x0e, 0x00, 0xee, 0x7c, 0x0e,
	0xe4, 0x4d, 0xec, 0xde, 0x9a, 0xf6, 0xe6, 0xd9, 0x65, 0xb7, 0x2a, 0xd0, 0x95, 0x5f, 0x9c, 0x00, 0xe6,
	0xec, 0xdf, 0x0c, 0xf2, 0xf7, 0xe2, 0xc8, 0x38, 0x03, 0xf7, 0x15, 0x6a, 0x9f, 0x06, 0x99, 0x4d, 0xa6,
	0x35, 0xd8, 0xa6, 0x53, 0xa2, 0xc4, 0x6c, 0x78, 0xe1, 0x1d, 0x94, 0xb6, 0x3f, 0x84, 0x81, 0xaf, 0xa7,
	0xd7, 0xc7, 0xcf, 0x8d, 0x55, 0x2d, 0xae, 0x97, 0xb7, 0xca, 0xe9, 0x8a, 0x89, 0xd0, 0xec, 0x67, 0xa1,
	0xc7, 0x70, 0x05, 0x21, 0x75, 0x27, 0xdb, 0x8a, 0x54, 0x0f, 0xbf, 0x05, 0x9e, 0x7f, 0xca, 0x75, 0x35,
	0x5c, 0x59, 0x94, 0xe5, 0x65, 0xca, 0xa2, 0x1e, 0x03, 0x90, 0x3c, 0x0b, 0x0c, 0x5c, 0xa1, 0xc0, 0x92,
	0x53, 0x30, 0x83, 0x08, 0x70, 0x53, 0xd8, 0x96, 0x04, 0xf9, 0x37, 0x27, 0xc9, 0xc4, 0x8a, 0x75, 0x32,
	0x9b, 0x53, 0x6b, 0xb1, 0x53, 0xa7, 0x45, 0x96, 0xd9, 0xfe, 0xc9, 0x0c, 0x25, 0xc1, 0xfc, 0x28, 0x1d,
	0x4c, 0x6f, 0x0e, 0x9a, 0xf6, 0xde, 0xe6, 0x6d, 0xc7, 0x29, 0x70, 0x57, 0x40, 0x9a, 0x47, 0xe5, 0x76,
	0x09, 0xff, 0x00, 0xcb, 0x3e, 0xdc, 0x9a, 0x51, 0xc6, 0x3f, 0x69, 0x4e, 0x97, 0x8d, 0xf4, 0xcb, 0x7e,
	0xda, 0xd7, 0x6d, 0xb1, 0x13, 0x73, 0x29, 0x4d, 0xd3, 0xed, 0xfa, 0xcd, 0x66, 0x35, 0x31, 0xf9, 0xd1,
	0xa7, 0x34, 0xea, 0x53, 0xdc, 0x51, 0xc2, 0x32, 0x0e, 0x50, 0x55, 0x82, 0x90, 0x48, 0xc6, 0x4f, 0xce,
	0x34, 0x9b, 0xdf, 0x0d, 0x81, 0xd9, 0xc8, 0x7b, 0x41, 0x5c, 0x99, 0x4f, 0xb5, 0x69, 0x54, 0x29, 0x34,
	0x3a, 0x73, 0xb2, 0xe1, 0xcd, 0x8c, 0xd8, 0x6d, 0xd4, 0x38, 0xda, 0x72, 0x94, 0xad, 0x7e, 0xae, 0x05,
	0x10, 0x12, 0x42, 0xb2, 0x4f, 0x2f, 0x9c, 0x1d, 0x6c, 0xca, 0xa5, 0x73, 0x68, 0x48, 0x89, 0x1a, 0x0a,
	0xcc, 0x23, 0x0f, 0xb1, 0x70, 0xf6, 0x7d, 0xb6, 0x6f, 0xbb, 0x60, 0xc5, 0x43, 0x50, 0xd3, 0x91, 0xe2,
	0xc0, 0xca, 0xe6, 0x61, 0x08, 0xe2, 0x35, 0xb5, 0x24, 0x03, 0xa6, 0xed, 0xee, 0x17, 0x77, 0xec, 0x2e,
	0xde, 0x1f, 0x81, 0x5b, 0x91, 0xb5, 0x17, 0x56, 0xda, 0x54, 0x13, 0x1e, 0xb7, 0x1d, 0x2e, 0xc4, 0x7c,
	0x91, 0x1a, 0x73, 0x19, 0x2c, 0xbd, 0x8f, 0x6c, 0x9f, 0x29, 0x57, 0xca, 0x4f, 0x9f, 0xb8, 0xf3, 0xac,
	0x66, 0xba, 0x31, 0x5e, 0xa0, 0xd2, 0x2e, 0x7a, 0x44, 0x9a, 0x15, 0x7a, 0x02, 0x25, 0xc1, 0x96, 0x8e,
	0x2e, 0xb4, 0xbf, 0xf4, 0x41, 0xfe, 0xd5, 0x03, 0xe4, 0x28, 0x79, 0x07, 0x41, 0xad, 0xd6, 0xdb, 0x79,
	0xfb, 0x65, 0x74, 0xbb, 0x44, 0x90, 0xb5, 0x3f, 0x11, 0xd1, 0xdf, 0x83, 0x24, 0x8c, 0x77, 0x99, 0x27,
	0x03, 0x3f, 0x0a, 0x07, 0xc2, 0x87, 0xc8, 0xf8, 0x23, 0x4c, 0x69, 0xca, 0x8c, 0x4d, 0x87, 0x42, 0x3d,
	0xa2, 0x8f, 0xc5, 0x0e, 0x47, 0xcf, 0x23, 0xc8, 0xf8, 0x9a, 0x56, 0xd4, 0x4a, 0xa9, 0xd5, 0x0c, 0xde,
	0x50, 0x95, 0x40, 0x51, 0x6b, 0xee, 0x93, 0xc1, 0xe4, 0x1e, 0xc7, 0xc1, 0xbb, 0x12, 0x9e, 0xe9, 0xb2,
	0xcc, 0x66, 0xd7, 0xdb, 0x98, 0xd5, 0x47, 0x63, 0x84, 0xcf, 0xaf, 0x1f, 0xc6, 0xbc, 0xb2, 0x3f, 0x37,
	0x6b, 0x24, 0x32, 0x9f, 0xb7, 0x1c, 0xab, 0xfc, 0xf5, 0x58, 0x42, 0x16, 0xe2, 0xd2, 0xdb, 0x68, 0x52,
	0xd6, 0xb2, 0x12, 0x94, 0xa4, 0x65, 0x4a, 0x24, 0xe0, 0x00, 0x3d, 0xc9, 0x3e, 0x35, 0xf7, 0xed, 0xcd,
	0x89, 0x51, 0xba, 0x2b, 0x14, 0x0d, 0xbd, 0xb7, 0x10, 0xd9, 0x95, 0x2b, 0xb3, 0x02, 0x39, 0x59, 0xc2,
	0x10, 0x12, 0x8c, 0x15, 0xa8, 0x8f, 0x44, 0xa5, 0x29, 0x2a, 0x38, 0x19, 0xf1, 0xf3, 0xa4, 0xfd, 0x4f,
	0xa5, 0xd3, 0xb1, 0x0e, 0xd1, 0xf7, 0x76, 0x45, 0x79, 0x37, 0x45, 0x3a, 0xd8, 0x98, 0xcd, 0x42, 0xb1,
	0x4f, 0x31, 0x3f, 0x0e, 0xa0, 0xca, 0x54, 0x32, 0xeb, 0x47, 0x92, 0xb9, 0x76, 0xc9, 0x0b, 0xe0, 0xac,
	0x64, 0x23, 0xd7, 0xdb, 0x40, 0x15, 0x94, 0xcd, 0xcf, 0x33, 0x11, 0x73, 0xc9, 0x1e, 0x8d, 0x57, 0x3c,
	0x0f, 0xf4, 0xb0, 0xe3, 0xb6, 0x2c, 0x29, 0x99, 0x4b, 0xa9, 0x3c, 0xa6, 0x5e, 0x54, 0xb5, 0x0e, 0xae,
	0x8f, 0x4a, 0x7f, 0xe9, 0x5f, 0x7b, 0x0d, 0x4a, 0x7b, 0x92, 0x2e, 0x71, 0xb2, 0xdb, 0xae, 0x92, 0xec,
	0x9b, 0x7f, 0x6a, 0x6a, 0x8a, 0xdc, 0xb4, 0xc7, 0x15, 0xca, 0xcd, 0x35, 0x6a, 0x99, 0x35, 0xe5, 0x00,
	0x9a, 0x32, 0x31, 0xc8, 0x06, 0x89, 0xf0, 0x95, 0x20, 0x80, 0x56, 0xbf, 0x72, 0x08, 0xcf, 0x1f, 0x07,
	0x07, 0x6d, 0x75, 0xb3, 0x54, 0xb3, 0xac, 0xda, 0x7d, 0xab, 0x26, 0xda, 0x6e, 0xe3, 0xab, 0x52, 0x92,
	0xa8, 0x6b, 0xaa, 0xaa, 0x69, 0x69, 0x89, 0x4d, 0x36, 0xa2, 0x96, 0x9d, 0x03, 0x81, 0x59, 0x2a, 0x40,
	0x4e, 0x73, 0x8f, 0x9f, 0x7d, 0x6b, 0xfa, 0xe6, 0xdc, 0x09, 0x70, 0xad, 0x3b, 0x7e, 0xca, 0xa4, 0xcb,
	0x52, 0x63, 0x5c, 0x4a, 0x72, 0x6c, 0xb5, 0x20, 0xe3, 0xbb, 0x1d, 0xae, 0x1c, 0x10, 0x7e, 0x95, 0x2d,
	0x61, 0x44, 0x7d, 0x03, 0x42, 0xbd, 0x6c, 0x4e, 0x66, 0x42, 0x49, 0x98, 0x19, 0x49, 0x60, 0xd0, 0x50,
	0x96, 0x51, 0xb1, 0x77, 0x63, 0x77, 0xb5, 0xb7, 0x7f, 0x93, 0x83, 0x14, 0xbc, 0x8c, 0xd5, 0x59, 0x25,
	0x4c, 0x67, 0xe7, 0xa8, 0x22, 0x2c, 0xa9, 0x29, 0x72, 0x02, 0x59, 0xd2, 0x59, 0x88, 0x20, 0x1d, 0xb4,
	0x8b, 0x7a, 0x41, 0x2e, 0x70, 0xbf, 0x85, 0xd6, 0x16, 0xde, 0x6e, 0x6d, 0x3d, 0xfb, 0x27, 0x78, 0xec,
	0x45, 0xc0, 0xa4, 0xd4, 0x80, 0x6d, 0xd9, 0x31, 0xe4, 0xaa, 0x43, 0x4d, 0x9c, 0x82, 0x95, 0xa8, 0x04,
	0xa5, 0xc6, 0xca, 0x48, 0x0a, 0x0b, 0x46, 0x4a, 0x48, 0xce, 0xb5, 0xdb, 0x97, 0xd3, 0xad, 0x62, 0xf4,
	0xdb, 0x69, 0x08, 0xb5, 0xb7, 0x8a, 0xeb, 0xb8, 0x9b, 0x43, 0x28, 0x99, 0x49, 0x83, 0x51, 0xa8, 0x34,
	0xfc, 0x39, 0x29, 0x48, 0xe4, 0x94, 0x15, 0xa1, 0x01, 0x4e, 0x12, 0x3f, 0xf8, 0x5a, 0x94, 0x70, 0x70,
	0x7f, 0x5d, 0x04, 0xb4, 0xbd, 0xe8, 0xd7, 0x7a, 0xa8, 0x94, 0x3b, 0x5a, 0xaf, 0x65, 0x5f, 0x57, 0x5d,
	0x3e, 0x99, 0x16, 0x98, 0xf2, 0x24, 0x52, 0x97, 0x3e, 0x5a, 0x5a, 0x1d, 0xb7, 0x79, 0x77, 0x1a, 0x41,
	0x51, 0x19, 0x09, 0x5a, 0x79, 0x60, 0x7a, 0x77, 0x0e, 0xbb, 0x28, 0x9c, 0x26, 0x6d, 0x14, 0xe4, 0xe6,
	0x8c, 0x75, 0x02, 0x02, 0xbd, 0xa7, 0xe8, 0xb3, 0x02, 0x38, 0x7d, 0x8f, 0xde, 0x3c, 0xd4, 0xb4, 0xca,
	0xa9, 0xc8, 0x02, 0x65, 0x20, 0x74, 0xe8, 0x50, 0x26, 0x1f, 0xbc, 0x6e, 0x03, 0xa5, 0xdc, 0x83, 0xb0,
	0x2c, 0x6e, 0x38, 0x6c, 0x10, 0xdc, 0x43, 0x8d, 0x2d, 0x4d, 0x3a, 0x85, 0x21, 0x68, 0x51, 0x4a, 0xd2,
	0xa1, 0x85, 0x25, 0x40, 0xe0, 0x82, 0x3d, 0x88, 0x3e, 0x08, 0xd4, 0xa7, 0xa9, 0x1b, 0x35, 0xab, 0xa3,
	0x6d, 0xe5, 0x54, 0x9a, 0x64, 0x2a, 0x6d, 0x04, 0xfe, 0x39, 0x95, 0x01, 0xf9, 0xbb, 0x7e, 0x03, 0xa9,
	0xfb, 0x71, 0xfc, 0xdf, 0xe1, 0xa4, 0xdf, 0x50, 0xe8, 0xb5, 0xce, 0xf0, 0xdc, 0x53, 0x6c, 0xea, 0xa4,
	0x2a, 0x85, 0x2a, 0xa2, 0xf2, 0x27, 0xb6, 0xf4, 0x37, 0x52, 0xe3, 0x5d, 0xc7, 0x50, 0x14, 0xea, 0x41,
	0x4f, 0x8f, 0x0e, 0x73, 0x24, 0x7e, 0xba, 0x98, 0xd4, 0x21, 0x33, 0x52, 0xa7, 0xca, 0xa6, 0xc9, 0x19,
	0x6a, 0x5b, 0x0e, 0x30, 0xe0, 0xfa, 0x56, 0x92, 0x93, 0xfe, 0x8e, 0x8c, 0xc1, 0x88, 0xa9, 0x4c, 0xc0,
	0x29, 0x25, 0xfa, 0x6a, 0xdc, 0x77, 0x00, 0xb1, 0xfd, 0x18, 0x79, 0x99, 0x82, 0x8a, 0x8a, 0x4c, 0x61,
	0xc4, 0x4b, 0x75, 0x91, 0xb1, 0xdc, 0x12, 0x1c, 0x79, 0x49, 0xfe, 0xc6, 0x3f, 0x67, 0x6e, 0x2f, 0xba,
	0x8d, 0xad, 0x59, 0xa0, 0x6e, 0x15, 0xb8, 0xa6, 0xff, 0x00, 0x15, 0x17, 0xb3, 0x3d, 0x80, 0xe0, 0xca,
	0x16, 0x95, 0xa3, 0x25, 0x0a, 0xc1, 0xcf, 0x15, 0x25, 0x45, 0x27, 0x07, 0x3e, 0x74, 0x9f, 0xaa, 0x75,
	0x47, 0xff, 0x00, 0x7c, 0x39, 0x48, 0xda, 0x17, 0x68, 0x29, 0xb5, 0xa9, 0xf7, 0x44, 0xc6, 0x69, 0xf5,
	0x7a, 0x82, 0xe5, 0xfe, 0x21, 0x5d, 0x95, 0x28, 0x72, 0x69, 0xa1, 0xc5, 0x3c, 0x7b, 0x84, 0x04, 0x73,
	0x56, 0x70, 0x17, 0xe9, 0xef, 0xae, 0x7d, 0x74, 0xdb, 0x78, 0xb5, 0x73, 0xed, 0xbc, 0x5a, 0x6b, 0x8f,
	0x05, 0x4d, 0xa0, 0xab, 0xf0, 0x2f, 0x24, 0x9f, 0xcd, 0xdb, 0xf2, 0x5a, 0x57, 0xdb, 0x8e, 0x53, 0xfe,
	0x1a, 0xab, 0x21, 0x6b, 0x69, 0x69, 0x75, 0xa7, 0x14, 0x85, 0xa1, 0x41, 0x49, 0x52, 0x4e, 0x14, 0x95,
	0x03, 0x90, 0x41, 0xf6, 0x20, 0xf9, 0xd7, 0xda, 0xac, 0xe6, 0x6e, 0x47, 0x1e, 0x2e, 0x45, 0x27, 0xd1,
	0xaa, 0xe3, 0x91, 0xfe, 0x87, 0x4f, 0x1d, 0xb1, 0x98, 0x99, 0x64, 0xba, 0xac, 0xca, 0x65, 0xe6, 0xab,
	0x4f, 0xf2, 0xe8, 0x1a, 0x55, 0x7f, 0x4a, 0xbe, 0xb6, 0x3a, 0x55, 0xd8, 0x82, 0x1c, 0x61, 0xa5, 0xd7,
	0x3d, 0x81, 0x2e, 0x6d, 0xa9, 0x6f, 0x5e, 0xb4, 0x98, 0x8a, 0x54, 0x6b, 0x75, 0x4e, 0x41, 0x96, 0x94,
	0x02, 0x7b, 0x51, 0xdd, 0xe1, 0xc1, 0x67, 0xe1, 0x29, 0x5a, 0x02, 0x49, 0xfa, 0xc6, 0x85, 0x9a, 0x71,
	0xed, 0xdf, 0x56, 0x76, 0x45, 0xc5, 0xb5, 0x15, 0x44, 0xee, 0x62, 0xa3, 0x7f, 0xcd, 0x51, 0xe9, 0xab,
	0x4c, 0xe8, 0x2f, 0x24, 0x71, 0xac, 0x23, 0x1c, 0x41, 0x69, 0x27, 0xc2, 0x8a, 0xc9, 0x01, 0x68, 0xf6,
	0x24, 0x9c, 0x71, 0xf2, 0x27, 0xf6, 0xdf, 0x44, 0xd5, 0x6b, 0xc6, 0xcd, 0xa7, 0x5d, 0x6f, 0x5c, 0xad,
	0x5b, 0x95, 0x5a, 0xaa, 0x55, 0x35, 0x74, 0x95, 0xc2, 0x2e, 0xb1, 0x15, 0xa7, 0x14, 0x54, 0xd3, 0x40,
	0xf3, 0x0b, 0x04, 0x20, 0x8c, 0xe7, 0x3f, 0x1e, 0xc7, 0x5b, 0x13, 0xa9, 0x68, 0x9d, 0x66, 0x06, 0x6e,
	0x58, 0x75, 0x95, 0xa5, 0xc8, 0xb0, 0x66, 0x61, 0x77, 0xb5, 0xf6, 0x6d, 0xec, 0x70, 0x66, 0x97, 0x9e,
	0x1a, 0x5b, 0x24, 0xa9, 0x74, 0xf8, 0x74, 0xc4, 0x35, 0x94, 0xa5, 0x4c, 0x48, 0x53, 0xba, 0x8b, 0x30,
	0x24, 0x81, 0xbe, 0xad, 0x99, 0x40, 0x16, 0xc1, 0x73, 0x4b, 0xbe, 0x8d, 0xb6, 0x52, 0x85, 0x5f, 0xb5,
	0xea, 0xf7, 0xad, 0xf7, 0x69, 0xc0, 0xa9, 0xc5, 0xa8, 0xbc, 0x88, 0xf4, 0xa4, 0xcf, 0x8a, 0x97, 0x47,
	0x6d, 0xae, 0x5d, 0xc7, 0x50, 0x14, 0x3c, 0x05, 0x2d, 0x5c, 0x72, 0x3d, 0x7b, 0x67, 0x5e, 0xf0, 0x7a,
	0x3b, 0xdb, 0xfd, 0xb6, 0xa7, 0xbf, 0x7a, 0xef, 0x25, 0xfa, 0xa9, 0xb4, 0x9a, 0x68, 0x0e, 0xbd, 0x1a,
	0x34, 0x75, 0x47, 0x69, 0xc3, 0x90, 0x12, 0x85, 0x2b, 0x92, 0x9c, 0x5f, 0x22, 0x40, 0x08, 0x46, 0x0a,
	0x89, 0xc6, 0xb5, 0xfb, 0x93, 0xd4, 0x65, 0x4a, 0xca, 0xdb, 0x67, 0xd5, 0x6a, 0x6c, 0xfd, 0xd9, 0x6f,
	0x36, 0xa6, 0x51, 0x0a, 0x91, 0x36, 0xa3, 0x4d, 0x6e, 0x3c, 0x38, 0xe9, 0x50, 0xe2, 0x95, 0x94, 0xa5,
	0x65, 0x4d, 0x94, 0x8c, 0x14, 0x21, 0x49, 0x4e, 0x4e, 0x07, 0xce, 0xbb, 0x28, 0x93, 0xa6, 0x53, 0x14,
	0xe6, 0xe6, 0x8c, 0x34, 0x82, 0x42, 0x7d, 0xc7, 0x87, 0x2c, 0xe0, 0x0d, 0xd9, 0xfb, 0xfd, 0x63, 0xcd,
	0x49, 0x53, 0x2e, 0xa2, 0x80, 0x99, 0x6d, 0x3e, 0xea, 0xd6, 0xa0, 0x0c, 0x4f, 0x60, 0xdc, 0x16, 0x49,
	0x2c, 0x49, 0xd8, 0x96, 0x16, 0x1c, 0xbe, 0x0b, 0x1d, 0x43, 0x1b, 0x59, 0xbd, 0xe1, 0xb8, 0xa0, 0x59,
	0x94, 0xa8, 0x54, 0xfa, 0x55, 0x39, 0xe4, 0x40, 0x6d, 0x98, 0x6d, 0x25, 0xb6, 0xbb, 0x8d, 0x20, 0x25,
	0xd5, 0x00, 0x9f, 0x19, 0x2e, 0x73, 0x04, 0xfe, 0x9a, 0x98, 0xd4, 0x26, 0xb3, 0x4d, 0xa7, 0xca, 0xa9,
	0x48, 0x38, 0x6a, 0x23, 0x0e, 0x3e, 0xe1, 0xfa, 0x50, 0x92, 0xa3, 0xfe, 0x86, 0xbe, 0x97, 0x1c, 0x71,
	0xd7, 0x14, 0xeb, 0xce, 0x29, 0xc7, 0x1c, 0x51, 0x5a, 0xd6, 0xb3, 0x95, 0x29, 0x44, 0xe4, 0x92, 0x7d,
	0xc9, 0x3e, 0x4e, 0xa5, 0x5d, 0x48, 0x5e, 0x4c, 0xda, 0xdb, 0x6f, 0x2e, 0x9c, 0x87, 0x82, 0x66, 0xd7,
	0x8f, 0xe0, 0x58, 0x48, 0x3e, 0x7b, 0x7e, 0x0b, 0xaa, 0xfb, 0x04, 0xfe, 0x5f, 0xba, 0xc6, 0x8c, 0xc1,
	0x86, 0xa9, 0xb4, 0xc0, 0x21, 0x01, 0xba, 0x8a, 0xd8, 0x76, 0x04, 0xb9, 0xfc, 0x18, 0x77, 0x99, 0x8e,
	0x8a, 0x76, 0x4c, 0x62, 0x44, 0x53, 0xf4, 0x61, 0xee, 0x7b, 0x90, 0x18, 0x79, 0x51, 0xfe, 0xce, 0x0b,
	0xdb, 0x53, 0xb9, 0x15, 0x0d, 0xb3, 0xba, 0x9a, 0xad, 0xc7, 0x42, 0x9f, 0x88, 0xea, 0x7b, 0x13, 0xa3,
	0x03, 0x8e, 0xf3, 0x24, 0x82, 0x71, 0xf0, 0xa0, 0x7c, 0xa4, 0xfc, 0x8f, 0x82, 0x74, 0xe5, 0xa0, 0x57,
	0xa9, 0x17, 0x3d, 0x22, 0x35, 0x7a, 0x85, 0x39, 0x12, 0xe0, 0xcb, 0x47, 0x36, 0x9d, 0x47, 0xbf, 0xc8,
	0x23, 0xd5, 0x2a, 0x07, 0xc1, 0x49, 0xf2, 0x0e, 0xb9, 0xcf, 0xad, 0x9e, 0xdb, 0xee, 0xb5, 0xd3, 0xb6,
	0x53, 0xd5, 0x22, 0x89, 0x21, 0x2e, 0xc4, 0x7c, 0x83, 0x26, 0x0b, 0xe4, 0x96, 0x5e, 0xfd, 0x70, 0x3c,
	0xa5, 0x5f, 0x0a, 0x1e, 0x7e, 0xe3, 0xc6, 0xa9, 0x95, 0x1d, 0x38, 0x26, 0xc3, 0xaf, 0x02, 0xd1, 0x47,
	0xe2, 0x87, 0x07, 0xe7, 0x83, 0xe0, 0xfc, 0x43, 0x28, 0x9a, 0xd9, 0x54, 0xea, 0x8e, 0x53, 0x36, 0x0a,
	0xa0, 0x28, 0xbd, 0xb7, 0x49, 0xe4, 0x72, 0x0f, 0x71, 0xe4, 0x5d, 0xc1, 0x7a, 0xfe, 0xbf, 0x1e, 0x75,
	0x57, 0xb6, 0xba, 0xa5, 0xdf, 0x2b, 0x5e, 0x2a, 0x20, 0xc5, 0xbd, 0x9c, 0x9d, 0x1d, 0xb0, 0x02, 0x51,
	0x52, 0x8e, 0xdc, 0xa2, 0x00, 0xf6, 0xe6, 0xa1, 0xcf, 0xf9, 0x51, 0xd1, 0x4e, 0xcd, 0xea, 0x47, 0x6d,
	0xee, 0x86, 0x9b, 0x6a, 0xa5, 0x3d, 0x54, 0x19, 0xaa, 0x00, 0x29, 0x99, 0xdf, 0xd2, 0xe5, 0xf4, 0xbc,
	0x07, 0x12, 0x3f, 0x77, 0x1d, 0x53, 0x21, 0x54, 0x69, 0xd5, 0x36, 0x44, 0x8a, 0x6d, 0x42, 0x34, 0xb6,
	0x88, 0xc8, 0x5b, 0x0f, 0x25, 0xc4, 0xff, 0x00, 0x29, 0x27, 0x53, 0x55, 0x42, 0x98, 0xc9, 0x96, 0x5c,
	0x2a, 0x19, 0xf8, 0x70, 0xfe, 0x45, 0x8e, 0x2e, 0x48, 0x8d, 0x25, 0xa9, 0xa1, 0x06, 0x30, 0xe3, 0x27,
	0xb0, 0x2c, 0x48, 0xf0, 0x6e, 0x3f, 0x06, 0x2d, 0xe8, 0xea, 0x6e, 0xfb, 0xba, 0xee, 0xcb, 0x62, 0x4e,
	0xe6, 0xd4, 0x99, 0x9f, 0x6f, 0x51, 0xeb, 0x51, 0x6a, 0x72, 0x20, 0xc5, 0x82, 0xdb, 0x69, 0x57, 0x6d,
	0x47, 0x0b, 0xc0, 0x1c, 0x96, 0x53, 0x92, 0xa0, 0x92, 0x71, 0x91, 0xf3, 0x8d, 0x26, 0xf7, 0xc7, 0x7e,
	0xb6, 0x72, 0x5e, 0xd0, 0x57, 0x61, 0xc1, 0xba, 0xe9, 0x35, 0xd7, 0xeb, 0x94, 0xe7, 0x62, 0x43, 0x85,
	0x19, 0xd0, 0xe3, 0xab, 0x71, 0xc4, 0xe1, 0x2a, 0x5a, 0x3d, 0x5b, 0x09, 0x24, 0x28, 0x95, 0x63, 0x1c,
	0x7e, 0x70, 0x35, 0xcf, 0x59, 0xd5, 0x1a, 0x75, 0x31, 0x93, 0x22, 0xa5, 0x50, 0x8d, 0x11, 0xa1, 0xe4,
	0xad, 0xf7, 0x92, 0xda, 0x7f, 0x95, 0x11, 0xa9, 0x9d, 0xe5, 0xd4, 0x8e, 0xdb, 0xda, 0xec, 0xba, 0x8a,
	0x6d, 0x44, 0xd7, 0xa7, 0x27, 0xc2, 0x59, 0x83, 0xfd, 0x3c, 0xfd, 0x4f, 0x11, 0xc7, 0x1f, 0xb7, 0x91,
	0xd6, 0xc4, 0xaa, 0x6b, 0x36, 0x8a, 0x88, 0x90, 0x60, 0xa0, 0xc5, 0x31, 0x2c, 0xe5, 0xcb, 0x59, 0xb7,
	0xd9, 0xbe, 0xed, 0x82, 0xf5, 0x0d, 0x3b, 0x4e, 0xc0, 0x8b, 0x03, 0x35, 0x99, 0x8a, 0x20, 0x26, 0x0d,
	0xc2, 0x52, 0x42, 0x75, 0x5d, 0xf6, 0x17, 0x77, 0xee, 0x03, 0x91, 0xe3, 0x14, 0x5a, 0xf5, 0x7a, 0x91,
	0x6c, 0x52, 0x24, 0xd7, 0x6b, 0xd3, 0x91, 0x12, 0x0c, 0x44, 0x73, 0x75, 0xe5, 0xff, 0x00, 0xa0, 0x07,
	0xaa, 0x94, 0x4f, 0x80, 0x91, 0xe4, 0x9d, 0x06, 0xf7, 0x63, 0x71, 0xe7, 0x6e, 0x6d, 0xd4, 0xed, 0x6d,
	0xf4, 0x2d, 0x88, 0x6c, 0xa7, 0xb1, 0x06, 0x31, 0x39, 0xec, 0xb2, 0x09, 0x23, 0x3f, 0x52, 0x8e, 0x54,
	0xa3, 0xf2, 0x71, 0xe8, 0x06, 0xbc, 0xdc, 0x7d, 0xd8, 0xba, 0xb7, 0x32, 0x72, 0x5e, 0xad, 0xc8, 0x4b,
	0x50, 0xd9, 0x51, 0x31, 0xa0, 0xb1, 0x90, 0xcb, 0x3f, 0xae, 0x3d, 0x54, 0xaf, 0xa9, 0x59, 0x3f, 0x18,
	0x1e, 0x35, 0x8b, 0xd3, 0x1a, 0x72, 0x9c, 0x12, 0x91, 0xd7, 0x8f, 0x78, 0xa7, 0xf1, 0x23, 0x81, 0xf3,
	0xc9, 0xf0, 0x3e, 0x66, 0xb5, 0xb5, 0x6c, 0xaa, 0x89, 0x43, 0x29, 0x94, 0x05, 0x30, 0x12, 0x5e, 0xfb,
	0xa8, 0xf2, 0x78, 0x03, 0xb0, 0xf2, 0x6e, 0xc0, 0x7f, 0xff, 0xd9, 0x27, 0x04, 0x0d, 0x83, 0x85, 0x74,
	0x65, 0x78, 0x74, 0x5f, 0x30, 0x2e, 0x74, 0x78, 0x74, 0x00, 0x8e, 0x74, 0x65, 0x78, 0x74, 0x5f, 0x30,
	0x2e, 0x74, 0x78, 0x74, 0x00, 0xc0, 0x22, 0x74, 0x65, 0x78, 0x74, 0x5f, 0x30, 0x2e, 0x74, 0x78, 0x74,
	0x00, 0x54, 0x65, 0x73, 0x74,
}

func getMRetrieveConfPayload(args mainFlags) []byte {

	from := mRetrieveConfFrom
	if args.Sender != "" {
		from = bytes.Join(
			[][]byte{
				// From + size + token address present +
				[]byte{0x80 + mms.FROM, byte(len(args.Sender)) + 12, mms.TOKEN_ADDRESS_PRESENT},
				// + sender +
				[]byte(args.Sender),
				// + "/TYPE=PLMN\0"
				[]byte{0x2f, 0x54, 0x59, 0x50, 0x45, 0x3d, 0x50, 0x4c, 0x4d, 0x4e, 0x00},
			},
			nil,
		)
	}

	transactionId := mRetrieveConfTransactionId
	if args.TransactionId != "" {
		transactionId = bytes.Join(
			[][]byte{
				// TransactionId +
				[]byte{0x80 + mms.X_MMS_TRANSACTION_ID},
				// + id string +
				[]byte(args.TransactionId),
				// + "\0"
				[]byte{0x00},
			},
			nil,
		)
	}

	return bytes.Join(
		[][]byte{
			mRetrieveConfType,
			transactionId,
			mRetrieveConfMMSVersion,
			mRetrieveConfMessageId,
			mRetrieveConfDate,
			from,
			mRetrieveConfContentType,
			mRetrieveConfAttachments,
		},
		nil,
	)
}
