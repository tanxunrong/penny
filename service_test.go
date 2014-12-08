package penny

import (
	"testing"
	"reflect"
)

func TestInit(t *testing.T) {
	slua := new(Slua)
	DefaultDock.AddService("slua", reflect.TypeOf(*slua), 100, 10)
}
