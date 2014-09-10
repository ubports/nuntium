/*
 * Copyright 2014 Canonical Ltd.
 *
 * Authors:
 * Sergio Schvezov: sergio.schvezov@canonical.com
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

package ofono

import (
	"errors"
	"fmt"

	"launchpad.net/go-dbus/v1"
	. "launchpad.net/gocheck"
)

type ContextTestSuite struct {
	modem    Modem
	contexts []OfonoContext
}

var _ = Suite(&ContextTestSuite{})

var proxy ProxyInfo

func makeGenericContextProperty(name, cType string, active, messageCenter, messageProxy bool) PropertiesType {
	p := make(PropertiesType)
	p["Name"] = dbus.Variant{name}
	p["Type"] = dbus.Variant{cType}
	p["Active"] = dbus.Variant{active}
	if messageCenter {
		p["MessageCenter"] = dbus.Variant{"http://messagecenter.com"}
	} else {
		p["MessageCenter"] = dbus.Variant{""}
	}
	if messageProxy {
		p["MessageProxy"] = dbus.Variant{proxy.String()}
	} else {
		p["MessageProxy"] = dbus.Variant{""}
	}
	return p
}

func (s *ContextTestSuite) SetUpSuite(c *C) {
}

func (s *ContextTestSuite) SetUpTest(c *C) {
	s.modem = Modem{}
	s.contexts = []OfonoContext{}
	proxy = ProxyInfo{
		Host: "4.4.4.4",
		Port: 9999,
	}
	getOfonoProps = func(conn *dbus.Connection, objectPath dbus.ObjectPath, destination, iface, method string) (oProps []OfonoContext, err error) {
		return s.contexts, nil
	}
}

func (s *ContextTestSuite) TestNoContext(c *C) {
	context, err := s.modem.GetMMSContext()
	c.Check(context.Properties, IsNil)
	c.Assert(err, DeepEquals, errors.New("No mms contexts found"))
}

func (s *ContextTestSuite) TestMMSOverInternet(c *C) {
	context1 := OfonoContext{
		ObjectPath: "/ril_0/context1",
		Properties: makeGenericContextProperty("Context1", contextTypeInternet, true, true, true),
	}
	s.contexts = append(s.contexts, context1)

	context, err := s.modem.GetMMSContext()
	c.Assert(err, IsNil)
	c.Check(context, DeepEquals, context1)
}

func (s *ContextTestSuite) TestMMSOverInactiveInternet(c *C) {
	context1 := OfonoContext{
		ObjectPath: "/ril_0/context1",
		Properties: makeGenericContextProperty("Context1", contextTypeInternet, false, true, true),
	}
	s.contexts = append(s.contexts, context1)

	context, err := s.modem.GetMMSContext()
	c.Check(context.Properties, IsNil)
	c.Assert(err, DeepEquals, errors.New("No mms contexts found"))
}

func (s *ContextTestSuite) TestMMSOverInternetNoProxy(c *C) {
	context1 := OfonoContext{
		ObjectPath: "/ril_0/context1",
		Properties: makeGenericContextProperty("Context1", contextTypeInternet, true, true, false),
	}
	s.contexts = append(s.contexts, context1)

	context, err := s.modem.GetMMSContext()
	c.Assert(err, IsNil)
	c.Check(context, DeepEquals, context1)
}

func (s *ContextTestSuite) TestMMSOverMMS(c *C) {
	context1 := OfonoContext{
		ObjectPath: "/ril_0/context1",
		Properties: makeGenericContextProperty("Context1", contextTypeInternet, true, false, false),
	}
	s.contexts = append(s.contexts, context1)

	context2 := OfonoContext{
		ObjectPath: "/ril_0/context2",
		Properties: makeGenericContextProperty("Context2", contextTypeMMS, false, true, true),
	}
	s.contexts = append(s.contexts, context2)

	context, err := s.modem.GetMMSContext()
	c.Assert(err, IsNil)
	c.Check(context, DeepEquals, context2)
}

func (s *ContextTestSuite) TestMMSOverMMSNoProxy(c *C) {
	context1 := OfonoContext{
		ObjectPath: "/ril_0/context1",
		Properties: makeGenericContextProperty("Context1", contextTypeInternet, true, false, false),
	}
	s.contexts = append(s.contexts, context1)

	context2 := OfonoContext{
		ObjectPath: "/ril_0/context2",
		Properties: makeGenericContextProperty("Context2", contextTypeMMS, false, true, false),
	}
	s.contexts = append(s.contexts, context2)

	context, err := s.modem.GetMMSContext()
	c.Assert(err, IsNil)
	c.Check(context, DeepEquals, context2)
}

func (s *ContextTestSuite) TestMMSPreferInternetOverMMS(c *C) {
	context1 := OfonoContext{
		ObjectPath: "/ril_0/context1",
		Properties: makeGenericContextProperty("Context1", contextTypeInternet, true, true, false),
	}
	s.contexts = append(s.contexts, context1)

	context2 := OfonoContext{
		ObjectPath: "/ril_0/context2",
		Properties: makeGenericContextProperty("Context2", contextTypeMMS, false, true, false),
	}
	s.contexts = append(s.contexts, context2)

	context, err := s.modem.GetMMSContext()
	c.Assert(err, IsNil)
	c.Check(context, DeepEquals, context1)
}

func (s *ContextTestSuite) TestGetProxy(c *C) {
	context := OfonoContext{
		ObjectPath: "/ril_0/context1",
		Properties: makeGenericContextProperty("Context1", contextTypeInternet, true, true, true),
	}

	p, err := context.GetProxy()
	c.Assert(err, IsNil)
	c.Check(p, DeepEquals, proxy)
}

func (s *ContextTestSuite) TestGetProxyNoProxy(c *C) {
	context := OfonoContext{
		ObjectPath: "/ril_0/context1",
		Properties: makeGenericContextProperty("Context1", contextTypeInternet, true, true, false),
	}

	p, err := context.GetProxy()
	c.Assert(err, IsNil)
	c.Check(p, DeepEquals, ProxyInfo{})
}

func (s *ContextTestSuite) TestGetProxyWithHTTP(c *C) {
	context := OfonoContext{
		ObjectPath: "/ril_0/context1",
		Properties: makeGenericContextProperty("Context1", contextTypeInternet, true, true, true),
	}
	context.Properties["MessageProxy"] = dbus.Variant{fmt.Sprintf("http://%s:%d", proxy.Host, proxy.Port)}

	p, err := context.GetProxy()
	c.Assert(err, IsNil)
	c.Check(p, DeepEquals, proxy)
}

func (s *ContextTestSuite) TestGetProxyNoPort(c *C) {
	context := OfonoContext{
		ObjectPath: "/ril_0/context1",
		Properties: makeGenericContextProperty("Context1", contextTypeInternet, true, true, true),
	}
	context.Properties["MessageProxy"] = dbus.Variant{fmt.Sprintf("http://%s", proxy.Host)}

	p, err := context.GetProxy()
	c.Assert(err, IsNil)
	c.Check(p, DeepEquals, ProxyInfo{Host: proxy.Host, Port: 80})
}
