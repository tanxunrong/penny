package penny

import (
	proto "./proto"
)

// any service should implement this interface.
type Service interface {
	Name() string
	Init() error
	Call(m proto.Msg) error
	Close() error
}

// start the default dock.
// path is the config file.
func Run(path string) *Dock {
	conf := ParseConfig(path)
	dock, err := NewDock(conf)
	if err != nil {
		panic(err)
	}

	return dock
}
