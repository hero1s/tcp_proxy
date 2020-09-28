package main

import (
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
)

var (
	pConfig    ProxyConfig
	pLog       *logrus.Logger
	configFile = flag.String("c", "./etc/conf.toml", "配置文件，默认etc/conf.toml")
)

func main() {
	openSignal()
	flag.Parse()
	fmt.Println("Start Proxy...")

	err := parseConfigFile(*configFile)
	if err != nil {
		fmt.Println("parse config fail:%v", err)
		return
	}
	// init logger server
	err = initLogger()
	if err != nil {
		fmt.Println("init logger fail:%v,path:%v", err, pConfig.Log.Path)
		return
	}

	// init Backend server
	initBackendSvrs()

	pidFile := fmt.Sprintf("pid.txt")
	ioutil.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0666)

	// init status service
	initStats()

	// init proxy service
	if pConfig.Websocket {
		init_websocket_proxy()
	} else {
		initProxy()
	}
}

// 自定义用户信号
var SigUser1 syscall.Signal

func openSignal() {
	SigUser1 = syscall.Signal(0xa)
	//创建监听退出chan
	c := make(chan os.Signal)
	//监听指定信号 ctrl+c kill
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, SigUser1)
	go func() {
		for s := range c {
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				fmt.Println("退出", s)
				exitFunc()
			case SigUser1:
				fmt.Println("usr1", s)
				exitFunc()
			default:
				fmt.Println("other", s)
			}
		}
	}()
}
func exitFunc() {
	os.Exit(0)
}
