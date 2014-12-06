package penny
import (
	"log"
	"os"
	proto "./proto"
)

type Slog struct {
	l *log.Logger
}

func (s *Slog) Name() string {
	return "slog"
}

func (s *Slog) Init() error {
	file,err := os.OpenFile("/tmp/penny.log",os.O_WRONLY | os.O_CREATE ,0644)
	if err != nil {
		return err
	}
	s.l = log.New(file,"[test]",log.Ldate | log.Ltime | log.Lshortfile)
	return nil
}

func (s *Slog) Close() error {
	s.l.Println("logger service shutdown")
	return nil
}

func (s *Slog) Call(m proto.Msg) error {
	return nil
}
