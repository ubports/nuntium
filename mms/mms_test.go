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

import . "launchpad.net/gocheck"

type MMSTestSuite struct{}

var _ = Suite(&MMSTestSuite{})

func (s *MMSTestSuite) TestNewMSendReq(c *C) {
	recipients := []string{"+11111", "+22222", "+33333"}
	recipientsStr := "+11111/TYPE=PLMN,+22222/TYPE=PLMN,+33333/TYPE=PLMN"
	mSendReq := NewMSendReq(recipients, []*Attachment{}, false)
	c.Check(mSendReq.To, Equals, recipientsStr)
	c.Check(mSendReq.ContentType, Equals, "application/vnd.wap.multipart.related")
	c.Check(mSendReq.Type, Equals, byte(TYPE_SEND_REQ))
}
