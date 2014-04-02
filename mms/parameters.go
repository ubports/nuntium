/*
 * Copyright 2014 Canonical Ltd.
 *
 * Authors:
 * Sergio Schvezov: sergio.schvezov@cannical.com
 *
 * This file is part of mms.
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

package mms

//Table 38 of Well-Known Parameter Assignments from OMA-WAP-MMS section 7.3
const (
	WSP_PARAMETER_TYPE_Q                  = 0x00 // Version 1.1 Q-value
	WSP_PARAMETER_TYPE_CHARSET            = 0x01 // Version 1.1 Well-known-charset
	WSP_PARAMETER_TYPE_LEVEL              = 0x02 // Version 1.1 Version-value
	WSP_PARAMETER_TYPE_TYPE               = 0x03 // Version 1.1 Integer-value
	WSP_PARAMETER_TYPE_NAME_DEFUNCT       = 0x05 // Version 1.1 Text-string
	WSP_PARAMETER_TYPE_FILENAME_DEFUNCT   = 0x06 // Version 1.1 Text-string
	WSP_PARAMETER_TYPE_DIFFERENCES        = 0x07 // Version 1.1 Field-name
	WSP_PARAMETER_TYPE_PADDING            = 0x08 // Version 1.1 Short-integer
	WSP_PARAMETER_TYPE_CONTENT_TYPE       = 0x09 // Version 1.2 Constrained-encoding
	WSP_PARAMETER_TYPE_START_DEFUNCT      = 0x0A // Version 1.2 Text-string
	WSP_PARAMETER_TYPE_START_INFO_DEFUNCT = 0x0B // Version 1.2 Text-string
	WSP_PARAMETER_TYPE_COMMENT_DEFUNCT    = 0x0C // Version 1.3 Text-string
	WSP_PARAMETER_TYPE_DOMAIN_DEFUNCT     = 0x0D // Version 1.3 Text-string
	WSP_PARAMETER_TYPE_MAX_AGE            = 0x0E // Version 1.3 Delta-seconds-value
	WSP_PARAMETER_TYPE_PATH_DEFUNCT       = 0x0F // Version 1.3 Text-string
	WSP_PARAMETER_TYPE_SECURE             = 0x10 // Version 1.3 No-value
	WSP_PARAMETER_TYPE_SEC                = 0x11 // Version 1.4 Short-integer
	WSP_PARAMETER_TYPE_MAC                = 0x12 // Version 1.4 Text-value
	WSP_PARAMETER_TYPE_CREATION_DATE      = 0x13 // Version 1.4 Date-value
	WSP_PARAMETER_TYPE_MODIFICATION_DATE  = 0x14 // Version 1.4 Date-value
	WSP_PARAMETER_TYPE_READ_DATE          = 0x15 // Version 1.4 Date-value
	WSP_PARAMETER_TYPE_SIZE               = 0x16 // Version 1.4 Integer-value
	WSP_PARAMETER_TYPE_NAME               = 0x17 // Version 1.4 Text-value
	WSP_PARAMETER_TYPE_FILENAME           = 0x18 // Version 1.4 Text-value
	WSP_PARAMETER_TYPE_START              = 0x19 // Version 1.4 Text-value
	WSP_PARAMETER_TYPE_START_INFO         = 0x1A // Version 1.4 Text-value
	WSP_PARAMETER_TYPE_COMMENT            = 0x1B // Version 1.4 Text-value
	WSP_PARAMETER_TYPE_DOMAIN             = 0x1C // Version 1.4 Text-value
	WSP_PARAMETER_TYPE_PATH               = 0x1D // Version 1.4 Text-value
	WSP_PARAMETER_TYPE_UNTYPED            = 0xFF // Version 1.4 Text-value
)

var CONTENT_TYPES []string = []string{
    "*/*", "text/*", "text/html", "text/plain",
    "text/x-hdml", "text/x-ttml", "text/x-vCalendar",
    "text/x-vCard", "text/vnd.wap.wml",
    "text/vnd.wap.wmlscript", "text/vnd.wap.wta-event",
    "multipart/*", "multipart/mixed", "multipart/form-data",
    "multipart/byterantes", "multipart/alternative",
    "application/*", "application/java-vm",
    "application/x-www-form-urlencoded",
    "application/x-hdmlc", "application/vnd.wap.wmlc",
    "application/vnd.wap.wmlscriptc",
    "application/vnd.wap.wta-eventc",
    "application/vnd.wap.uaprof",
    "application/vnd.wap.wtls-ca-certificate",
    "application/vnd.wap.wtls-user-certificate",
    "application/x-x509-ca-cert",
    "application/x-x509-user-cert",
    "image/*", "image/gif", "image/jpeg", "image/tiff",
    "image/png", "image/vnd.wap.wbmp",
    "application/vnd.wap.multipart.*",
    "application/vnd.wap.multipart.mixed",
    "application/vnd.wap.multipart.form-data",
    "application/vnd.wap.multipart.byteranges",
    "application/vnd.wap.multipart.alternative",
    "application/xml", "text/xml",
    "application/vnd.wap.wbxml",
    "application/x-x968-cross-cert",
    "application/x-x968-ca-cert",
    "application/x-x968-user-cert",
    "text/vnd.wap.si",
    "application/vnd.wap.sic",
    "text/vnd.wap.sl",
    "application/vnd.wap.slc",
    "text/vnd.wap.co",
    "application/vnd.wap.coc",
    "application/vnd.wap.multipart.related",
    "application/vnd.wap.sia",
    "text/vnd.wap.connectivity-xml",
    "application/vnd.wap.connectivity-wbxml",
    "application/pkcs7-mime",
    "application/vnd.wap.hashed-certificate",
    "application/vnd.wap.signed-certificate",
    "application/vnd.wap.cert-response",
    "application/xhtml+xml",
    "application/wml+xml",
    "text/css",
    "application/vnd.wap.mms-message",
    "application/vnd.wap.rollover-certificate",
    "application/vnd.wap.locc+wbxml",
    "application/vnd.wap.loc+xml",
    "application/vnd.syncml.dm+wbxml",
    "application/vnd.syncml.dm+xml",
    "application/vnd.syncml.notification",
    "application/vnd.wap.xhtml+xml",
    "application/vnd.wv.csp.cir",
    "application/vnd.oma.dd+xml",
    "application/vnd.oma.drm.message",
    "application/vnd.oma.drm.content",
    "application/vnd.oma.drm.rights+xml",
    "application/vnd.oma.drm.rights+wbxml",
}

var CHARSETS map[uint64]string = map[uint64]string {
    0x07EA: "big5",
    0x03E8: "iso-10646-ucs-2",
    0x04: "iso-8859-1",
    0x05: "iso-8859-2",
    0x06: "iso-8859-3",
    0x07: "iso-8859-4",
    0x08: "iso-8859-5",
    0x09: "iso-8859-6",
    0x0A: "iso-8859-7",
    0x0B: "iso-8859-8",
    0x0C: "iso-8859-9",
    0x11: "shift_JIS",
    0x03: "us-ascii",
    0x6A: "utf-8",
}


