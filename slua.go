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

func (s *Slua) Init() error {
	s.L = luar.Init()
	return nil
}

func (s *Slua) Close() error {
	s.L.Close()
	return nil
}

func (s *Slua) Call(m Msg) {
	s.L.DoString(string(m.data))
}
