package controller

import (
	"GraduationProjection/fsm"
	"fmt"
	"sync"
	"time"
)

type fsmNode struct {
	fsmNodeConfig

	tickMu  *sync.Mutex
	ticker  *time.Ticker
	stopped chan struct{}
	done    chan struct{}
}

// fsmNodeConfig包含FSM相关组件，不包含fsmNode的相关组
type fsmNodeConfig struct {
	fsm.FSM
	heartbeat time.Duration // 两次时钟触发的间隔
}

func newFsmNode(fsmNodeCfg fsmNodeConfig) *fsmNode {

	f := &fsmNode{
		fsmNodeConfig: fsmNodeCfg,
		tickMu:        new(sync.Mutex),
		stopped:       make(chan struct{}),
		done:          make(chan struct{}),
	}

	if fsmNodeCfg.heartbeat == 0 {
		f.ticker = &time.Ticker{}
	} else {
		f.ticker = time.NewTicker(fsmNodeCfg.heartbeat)
	}
	return f
}

func (f *fsmNode) start() {
	// fsmNode开始监听各个通道
	go f.run()
}

func (f *fsmNode) run() {
	for {
		select {
		// 当收到时钟发出的信号时
		case <-f.ticker.C:
			f.tick()
			fmt.Println("发出时钟触发")
		}
	}
}

// 向底层状态机触发时钟信号
func (f *fsmNode) tick() {
	f.tickMu.Lock()
	f.FSM.Tick()
	f.tickMu.Unlock()
}
