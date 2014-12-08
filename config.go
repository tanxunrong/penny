package penny

import (
	"github.com/BurntSushi/toml"
	"io/ioutil"
)

// config for dock.
type Config struct {
	Machines []string
	Log      string
	Addr     string
}

// parse toml file and return dock config.
func ParseConfig(file string) *Config {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	conf := new(Config)
	if _, err = toml.Decode(string(content), conf); err != nil {
		panic(err)
	}
	if len(conf.Machines) == 0 {
		panic("etcd cluster addrs invalid")
	}
	if len(conf.Log) == 0 {
		panic("invalid log path")
	}
	if len(conf.Addr) == 0 {
		panic("invalid local bind addr")
	}
	return conf
}
