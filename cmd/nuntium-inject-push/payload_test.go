package main

import (
	"reflect"
	"testing"
)

func TestMRetrieveConfSplitted(t *testing.T) {
	if !reflect.DeepEqual(mRetrieveConf, mRetrieveConfSplited) {
		t.Errorf("Not equal")
	}
}
