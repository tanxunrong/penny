package main

import (
	"reflect"
	penny "../"
)
func main() {
	dock := penny.Run("/home/tanxr/workspace/penny/conf.toml")
	slua := new(penny.Slua)
	dock.AddService("slua", reflect.TypeOf(*slua), 100, 10)

	slog := new(penny.Slog)
	dock.AddService("slog", reflect.TypeOf(*slog), 10, 1)
	dock.Dispatch()
}
