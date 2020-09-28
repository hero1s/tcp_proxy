package main

import (
	"net"
	"time"
)

func initProxy() {

	pLog.Infof("Proxying %s -> %s\n", pConfig.Bind, pConfig.Backend)

	server, err := net.Listen("tcp", pConfig.Bind)
	if err != nil {
		pLog.Fatal(err)
	}

	waitQueue := make(chan net.Conn, pConfig.WaitQueueLen)
	availPools := make(chan bool, pConfig.MaxConn)
	for i := 0; i < pConfig.MaxConn; i++ {
		availPools <- true
	}

	go loop(waitQueue, availPools)

	for {
		connection, err := server.Accept()
		if err != nil {
			pLog.Error(err)
		} else {
			pLog.Infof("Received connection from %s.\n", connection.RemoteAddr())
			waitQueue <- connection
		}
	}
}

func loop(waitQueue chan net.Conn, availPools chan bool) {
	for connection := range waitQueue {
		<-availPools
		go func(connection net.Conn) {
			handleConnection(connection)
			availPools <- true
			pLog.Infof("Closed connection from %s.\n", connection.RemoteAddr())
		}(connection)
	}
}

func handleConnection(connection net.Conn) {
	defer connection.Close()

	bksvr, ok := getBackendSvr(connection.RemoteAddr().String())
	if !ok {
		return
	}
	remote, err := net.Dial("tcp", bksvr.svrStr)
	if err != nil {
		return
	}
	onConnected(err,bksvr)

	//等待双向连接完成
	complete := make(chan bool, 2)
	oneSide := make(chan bool, 1)
	otherSide := make(chan bool, 1)
	go pass(connection, remote, complete, oneSide, otherSide)
	go pass(remote, connection, complete, otherSide, oneSide)
	<-complete
	<-complete
	remote.Close()
}

// copy Content two-way
func pass(from net.Conn, to net.Conn, complete chan bool, oneSide chan bool, otherSide chan bool) {
	var err error
	var read int
	bytes := make([]byte, 256)

	for {
		select {

		case <-otherSide:
			complete <- true
			return

		default:

			from.SetReadDeadline(time.Now().Add(time.Duration(pConfig.Timeout) * time.Second))
			read, err = from.Read(bytes)
			if err != nil {
				complete <- true
				oneSide <- true
				return
			}

			to.SetWriteDeadline(time.Now().Add(time.Duration(pConfig.Timeout) * time.Second))
			_, err = to.Write(bytes[:read])
			if err != nil {
				complete <- true
				oneSide <- true
				return
			}
		}
	}
}


