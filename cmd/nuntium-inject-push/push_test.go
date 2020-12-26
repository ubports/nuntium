package main

import (
	"reflect"
	"testing"
)

func TestMNotificationIndSplitted(t *testing.T) {
	if !reflect.DeepEqual(mNotificationInd, mNotificationIndSplitted) {
		t.Errorf("Not equal")
	}
}
