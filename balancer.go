package main

import (
	//"fmt"
	"math/rand"
	"stathat.com/c/consistent"
	"time"
)

// BackendSvr Type
type BackendSvr struct {
	svrStr          string
	isUp            bool // is Up or Down
	failTimes       int  // 累计失败次数
	latelyFailTimes int  // 最近失败次数
}

var (
	pConsisthash *consistent.Consistent
	pBackendSvrs map[string]*BackendSvr
)

func initBackendSvrs() {
	pConsisthash = consistent.New()
	pBackendSvrs = make(map[string]*BackendSvr)

	for _, svr := range pConfig.Backend {
		pConsisthash.Add(svr)
		pBackendSvrs[svr] = &BackendSvr{
			svrStr:          svr,
			isUp:            true,
			failTimes:       0,
			latelyFailTimes: 0,
		}
	}
	go checkBackendSvrs()
}

func reinitBackendSvrs() {
	pLog.Errorln("all server is stop,reinit backend svrs")
	pConsisthash = consistent.New()
	pBackendSvrs = make(map[string]*BackendSvr)

	for _, svr := range pConfig.Backend {
		pConsisthash.Add(svr)
		pBackendSvrs[svr] = &BackendSvr{
			svrStr:          svr,
			isUp:            true,
			failTimes:       0,
			latelyFailTimes: 0,
		}
	}

}

func getBackendSvr(remoteAddr string) (*BackendSvr, bool) {
	svr, _ := pConsisthash.Get(remoteAddr)
	bksvr, ok := pBackendSvrs[svr]
	return bksvr, ok
}

func onConnected(err error, bksvr *BackendSvr) {
	if err != nil {
		pLog.Error(err)
		bksvr.failTimes++
		bksvr.latelyFailTimes++
	} else {
		bksvr.latelyFailTimes = 0
	}
}

func checkBackendSvrs() {
	// scheduler every 10 seconds
	rand.Seed(time.Now().UnixNano())
	t := time.Tick(time.Duration(10)*time.Second + time.Duration(rand.Intn(100))*time.Millisecond*100)

	for _ = range t {
		for _, v := range pBackendSvrs {
			if v.latelyFailTimes >= pConfig.FailOver && v.isUp == true {
				v.isUp = false
				pConsisthash.Remove(v.svrStr)
			}
		}
	}
	if len(pConsisthash.Members()) < 1 {
		reinitBackendSvrs()
	}
}
