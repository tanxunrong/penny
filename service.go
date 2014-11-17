package penny

import (
	"sync"
)

type Service interface {
	Name() string
	Open() Service
	Call(m Msg) error
	Close()
}

type Msg struct {
	source uint32
	session int
	data []byte
}

type Center struct {
	storage map[string]Service
	instances []*Service
	mutex sync.Mutex
	id uint16
}
