package penny

import (
	lua "github.com/aarzilli/golua/lua"
	luar "github.com/stevedonovan/luar"
)

type Slua struct {
	L *lua.State
}

func (s *Slua) Name() string {
	return "slua"
}

func (s *Slua) Open() Slua {
	l := luar.Init()
	return Slua{L:l}
}

func (s *Slua) Close() error {
	s.L.Close()
	return nil
}

func (s *Slua) Call(m Msg) {
	s.L.DoString(string(m.data))
}
