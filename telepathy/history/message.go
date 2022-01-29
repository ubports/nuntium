package history

import (
	"fmt"
	"launchpad.net/go-dbus"
)

type Message map[string]dbus.Variant

func (m Message) Exists() bool {
	return m != nil
}

var ErrorNonExistentMessage = fmt.Errorf("message doesn't exist")

type ErrorMessagePropertyMissing string

func (e ErrorMessagePropertyMissing) Error() string {
	return fmt.Sprintf("Message proprety missing: %s", string(e))
}

type ErrorMessagePropertyType struct {
	property       string
	wantType, have interface{}
}

func (e ErrorMessagePropertyType) Error() string {
	return fmt.Sprintf("Message property \"%s\" type is %T, want %T", e.property, e.have, e.wantType)
}

func (m Message) IsNew() (bool, error) {
	if !m.Exists() {
		return false, ErrorNonExistentMessage
	}
	v, ok := m["newEvent"]
	if !ok {
		return false, ErrorMessagePropertyMissing("newEvent")
	}

	newEvent, ok := v.Value.(bool)
	if !ok {
		return false, ErrorMessagePropertyType{"newEvent", false, v.Value}
	}

	return newEvent, nil
}
