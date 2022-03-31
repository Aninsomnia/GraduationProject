package controller

import (
	"GraduationProjection/fsm"
	"sync"
	"time"
)

type fsmNode struct {
	fsm.FSM
	tickMu  *sync.Mutex
	ticker  *time.Ticker
	stopped chan struct{}
	done    chan struct{}
}

// TODO:采用bootstrapped方式传入cfg
func newFsmNode(c *fsm.Config) *fsmNode {
	return &fsmNode{
		FSM:     fsm.StartFsm(c),
		tickMu:  new(sync.Mutex),
		stopped: make(chan struct{}),
		done:    make(chan struct{}),
		// TODO:初始化ticker
	}
}

func (f *fsmNode) start() {
	go f.run()
}

func (f *fsmNode) run() {
	for {
		select {
		// 当收到时钟发出的信号时
		case <-f.ticker.C:
			f.tick()
		}
	}
}

// 向底层状态机触发时钟信号
func (f *fsmNode) tick() {
	f.tickMu.Lock()
	f.FSM.Tick()
	f.tickMu.Unlock()
}
