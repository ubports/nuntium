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

package ofono

//These are the WSP assigned numbers from Table 34. PDU Type Assignments -
//Appendix A Assigned Numbers in WAP-230-WSP
const (
	CONNECT        PDU = 0x01
	CONNECT_REPLY  PDU = 0x02
	REDIRECT       PDU = 0x03
	REPLY          PDU = 0x04
	DISCONNECT     PDU = 0x05
	PUSH           PDU = 0x06
	CONFIRMED_PUSH PDU = 0x07
	SUSPEND        PDU = 0x08
	RESUME         PDU = 0x09
	GET            PDU = 0x40
	POST           PDU = 0x60
)

//These are the WSP assigned numbers from Table 38 . Well-Known Parameter
//Assignments - Appendix A Assigned Numbers in WAP-230-WSP
const (
	WSP_PARAMETER_TYPE_Q                  = 0x00
	WSP_PARAMETER_TYPE_CHARSET            = 0x01
	WSP_PARAMETER_TYPE_LEVEL              = 0x02
	WSP_PARAMETER_TYPE_TYPE               = 0x03
	WSP_PARAMETER_TYPE_NAME_DEFUNCT       = 0x05
	WSP_PARAMETER_TYPE_FILENAME_DEFUNCT   = 0x06
	WSP_PARAMETER_TYPE_DIFFERENCES        = 0x07
	WSP_PARAMETER_TYPE_PADDING            = 0x08
	WSP_PARAMETER_TYPE_CONTENT_TYPE       = 0x09
	WSP_PARAMETER_TYPE_START_DEFUNCT      = 0x0A
	WSP_PARAMETER_TYPE_START_INFO_DEFUNCT = 0x0B
	WSP_PARAMETER_TYPE_COMMENT_DEFUNCT    = 0x0C
	WSP_PARAMETER_TYPE_DOMAIN_DEFUNCT     = 0x0D
	WSP_PARAMETER_TYPE_MAX_AGE            = 0x0E
	WSP_PARAMETER_TYPE_PATH_DEFUNCT       = 0x0F
	WSP_PARAMETER_TYPE_SECURE             = 0x10
	WSP_PARAMETER_TYPE_SEC                = 0x11
	WSP_PARAMETER_TYPE_MAC                = 0x12
	WSP_PARAMETER_TYPE_CREATION_DATE      = 0x13
	WSP_PARAMETER_TYPE_MODIFICATION_DATE  = 0x14
	WSP_PARAMETER_TYPE_READ_DATE          = 0x15
	WSP_PARAMETER_TYPE_SIZE               = 0x16
	WSP_PARAMETER_TYPE_NAME               = 0x17
	WSP_PARAMETER_TYPE_FILENAME           = 0x18
	WSP_PARAMETER_TYPE_START              = 0x19
	WSP_PARAMETER_TYPE_START_INFO         = 0x1A
	WSP_PARAMETER_TYPE_COMMENT            = 0x1B
	WSP_PARAMETER_TYPE_DOMAIN             = 0x1C
	WSP_PARAMETER_TYPE_PATH               = 0x1D
	WSP_PARAMETER_TYPE_UNTYPED            = 0xFF
)

//These are the WSP assigned numbers from Table 39 . Header Field Name
//Assignments - Appendix A Assigned Numbers in WAP-230-WSP
const (
	ACCEPT                = 0x00
	ACCEPT_CHARSET_1      = 0x01
	ACCEPT_ENCODING_1     = 0x02
	ACCEPT_LANGUAGE       = 0x03
	ACCEPT_RANGES         = 0x04
	AGE                   = 0x05
	ALLOW                 = 0x06
	AUTHORIZATION         = 0x07
	CACHE_CONTROL_1       = 0x08
	CONNECTION            = 0x09
	CONTENT_BASE          = 0x0A
	CONTENT_ENCODING      = 0x0B
	CONTENT_LANGUAGE      = 0x0C
	CONTENT_LENGTH        = 0x0D
	CONTENT_LOCATION      = 0x0E
	CONTENT_MD5           = 0x0F
	CONTENT_RANGE_1       = 0x10
	CONTENT_TYPE          = 0x11
	DATE                  = 0x12
	ETAG                  = 0x13
	EXPIRES               = 0x14
	FROM                  = 0x15
	HOST                  = 0x16
	IF_MODIFIED_SINCE     = 0x17
	IF_MATCH              = 0x18
	IF_NONE_MATCH         = 0x19
	IF_RANGE              = 0x1A
	IF_UNMODIFIED_SINCE   = 0x1B
	LOCATION              = 0x1C
	LAST_MODIFIED         = 0x1D
	MAX_FORWARDS          = 0x1E
	PRAGMA                = 0x1F
	PROXY_AUTHENTICATE    = 0x20
	PROXY_AUTHORIZATION   = 0x21
	PUBLIC                = 0x22
	RANGE                 = 0x23
	REFERER               = 0x24
	RETRY_AFTER           = 0x25
	SERVER                = 0x26
	TRANSFER_ENCODING     = 0x27
	UPGRADE               = 0x28
	USER_AGENT            = 0x29
	VARY                  = 0x2A
	VIA                   = 0x2B
	WARNING               = 0x2C
	WWW_AUTHENTICATE      = 0x2D
	CONTENT_DISPOSITION_1 = 0x2E
	X_WAP_APPLICATION_ID  = 0x2F
	X_WAP_CONTENT_URI     = 0x30
	X_WAP_INITIATOR_URI   = 0x31
	ACCEPT_APPLICATION    = 0x32
	BEARER_INDICATION     = 0x33
	PUSH_FLAG             = 0x34
	PROFILE               = 0x35
	PROFILE_DIFF          = 0x36
	PROFILE_WARNING_1     = 0x37
	EXPECT                = 0x38
	TE                    = 0x39
	TRAILER               = 0x3A
	ACCEPT_CHARSET        = 0x3B
	ACCEPT_ENCODING       = 0x3C
	CACHE_CONTROL_2       = 0x3D
	CONTENT_RANGE         = 0x3E
	X_WAP_TOD             = 0x3F
	CONTENT_ID            = 0x40
	SET_COOKIE            = 0x41
	COOKIE                = 0x42
	ENCODING_VERSION      = 0x43
	PROFILE_WARNING       = 0x44
	CONTENT_DISPOSITION   = 0x45
	X_WAP_SECURITY        = 0x46
	CACHE_CONTROL         = 0x47
)
