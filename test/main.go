package main

import (
	"reflect"
	penny "../"
)
func main() {
	penny.Run("/home/tanxr/workspace/penny/conf.toml");
	slua := new(penny.Slua)
	penny.DefaultDock.AddService("slua", reflect.TypeOf(*slua), 100, 10)

	slog := new(penny.Slog)
	penny.DefaultDock.AddService("slog", reflect.TypeOf(*slog), 10, 1)
}
