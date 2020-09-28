package main

import (
	"golang.org/x/net/websocket"
	"io"
	"net"
	"net/http"
)

var binaryMode = true

func copyWorker(dst io.Writer, src io.Reader, doneCh chan<- bool) {
	io.Copy(dst, src)
	doneCh <- true
}

func relayHandler(ws *websocket.Conn) {
	bksvr, ok := getBackendSvr(ws.RemoteAddr().String())
	if !ok {
		return
	}
	conn, err := net.Dial("tcp", bksvr.svrStr)
	if err != nil {
		return
	}
	onConnected(err,bksvr)
	if binaryMode {
		ws.PayloadType = websocket.BinaryFrame
	}
	doneCh := make(chan bool)
	go copyWorker(conn, ws, doneCh)
	go copyWorker(ws, conn, doneCh)
	<-doneCh
	conn.Close()
	ws.Close()
	<-doneCh
}

func init_websocket_proxy() {
	pLog.Infof("websocket Listening on %s ---- %s", pConfig.Bind, pConfig.Backend)
	http.Handle("/", websocket.Handler(relayHandler))
	var err error
	if pConfig.Tlscert != "" && pConfig.Tlskey != "" {
		err = http.ListenAndServeTLS(pConfig.Bind, pConfig.Tlscert, pConfig.Tlskey, nil)
	} else {
		err = http.ListenAndServe(pConfig.Bind, nil)
	}
	if err != nil {
		pLog.Fatal(err)
	}

}
