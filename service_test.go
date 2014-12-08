package penny

import (
	"testing"
	"reflect"
)

func TestInit(t *testing.T) {
	Run("/home/tanxr/workspace/penny/conf.toml");
	slua := new(Slua)
	DefaultDock.AddService("slua", reflect.TypeOf(*slua), 100, 10)

	slog := new(Slog)
	DefaultDock.AddService("slog", reflect.TypeOf(*slog), 10, 1)
}
