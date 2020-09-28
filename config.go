package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
)

// ProxyConfig Type
type ProxyConfig struct {
	Bind         string    `toml:"bind"`
	WaitQueueLen int       `toml:"wait_queue_len"`
	MaxConn      int       `toml:"max_conn"`
	Timeout      int       `toml:"timeout"`
	FailOver     int       `toml:"failover"`
	Backend      []string  `toml:"backend"`
	Log          LogConfig `toml:"log"`
	Stats        string    `toml:"stats"`
	Websocket    bool      `toml:"websocket"`
	Tlscert      string    `toml:"tlscert"`
	Tlskey       string    `toml:"tlskey"`
}

// LogConfig Type
type LogConfig struct {
	Level string `toml:"level"`
	Path  string `toml:"path"`
}

func parseConfigFile(filepath string) error {
	if _, err := toml.DecodeFile(filepath, &pConfig); err != nil {
		fmt.Println("config error:%v", err)
		return err
	}
	return nil
}
