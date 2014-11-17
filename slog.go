package penny
import (
	"log"
	"os"
)

type Slog struct {
	l *log.Logger
}

func (s *Slog) Name() string {
	return "logger"
}

func (s *Slog) Open() Slog {
	file,err := os.OpenFile("/tmp/penny.log",os.O_WRONLY | os.O_CREATE ,0644)
	if err != nil {
		panic(err)
	}
	l := log.New(file,"[test]",log.Ldate | log.Ltime | log.Lshortfile)
	return Slog{l:l}
}

func (s *Slog) Close() {
	s.l.Println("logger service shutdown")
}

func (s *Slog) Call(m Msg) error {
	s.l.Println(string(m.data))
	return nil
}
