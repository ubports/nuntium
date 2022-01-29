package history

import (
	"launchpad.net/go-dbus"
	"reflect"
	"testing"
)

func TestMessage_Exists(t *testing.T) {
	testCases := []struct {
		name string
		m    Message
		want bool
	}{
		{},
		{"nil", Message(nil), false},
		{"empty", Message{}, true},
		{"not empty", Message{"aaa": dbus.Variant{nil}}, true},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e := tc.m.Exists()
			if e != tc.want {
				t.Errorf("%#v.Exists() = %v, want %v", tc.m, e, tc.want)
			}
		})
	}
}

func TestMessage_IsNew(t *testing.T) {
	testCases := []struct {
		name string
		m    Message
		want bool
		err  error
	}{
		{"nil", Message(nil), false, ErrorNonExistentMessage},
		{"empty", Message{}, false, ErrorMessagePropertyMissing("newEvent")},
		{"missing", Message{"aaa": dbus.Variant{nil}}, false, ErrorMessagePropertyMissing("newEvent")},
		{"wrong type nil", Message{"newEvent": dbus.Variant{nil}}, false, ErrorMessagePropertyType{"newEvent", false, nil}},
		{"wrong type int", Message{"newEvent": dbus.Variant{10}}, false, ErrorMessagePropertyType{"newEvent", false, 10}},
		{"true", Message{"newEvent": dbus.Variant{true}}, true, nil},
		{"false", Message{"newEvent": dbus.Variant{false}}, false, nil},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isnew, err := tc.m.IsNew()
			if isnew != tc.want || !reflect.DeepEqual(err, tc.err) {
				t.Errorf("%#v.IsNew() = %v, %#v, want %v, %#v", tc.m, isnew, err, tc.want, tc.err)
			}
		})
	}
}
