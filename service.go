package penny

import (
	"sync"
	luar "github.com/stevedonovan/luar"
)
type Service interface {
	func Name() string
	func Open() *Service
	func Init(param string) bool
	func Close()
}

type Msg struct {
	source uint32
	session int
	data []byte
}

type Center struct {
	storage map[string]Service
}

type Msgqueue struct {
	ch chan Msg
}
