package controller

import (
	"GraduationProjection/fsm"
	"GraduationProjection/fsm/fsmpb"
	"sync"
	"time"

	"go.uber.org/zap"
)

type fsmNode struct {
	fsmNodeConfig

	logger  *zap.Logger
	tickMu  *sync.Mutex
	ticker  *time.Ticker
	stopped chan struct{}
	done    chan struct{}
}

// fsmNodeConfig包含FSM相关组件，不包含fsmNode的相关组
type fsmNodeConfig struct {
	logger *zap.Logger
	fsm.FSM
	heartbeat time.Duration // 两次时钟触发的间隔
}

func newFsmNode(fsmNodeCfg fsmNodeConfig) *fsmNode {

	f := &fsmNode{
		fsmNodeConfig: fsmNodeCfg,

		logger:  fsmNodeCfg.logger,
		tickMu:  new(sync.Mutex),
		stopped: make(chan struct{}),
		done:    make(chan struct{}),
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
		case rd := <-f.FSM.Ready():
			f.logger.Info("recieve Ready")
			err := f.handleReady(rd)
			if err != nil {
				f.logger.Warn("handle Ready error")
			}
			f.FSM.Advance()
		}

	}
}

// 向底层状态机触发时钟信号
func (f *fsmNode) tick() {
	f.tickMu.Lock()
	f.FSM.Tick()
	f.tickMu.Unlock()
}

func (f *fsmNode) handleReady(rd fsm.Ready) error {

	for _, msg := range rd.Messages {
		//TODE: 处理msg
		switch msg.Type {
		case fsmpb.MsgArbitRequest:
			// TODO:发送仲裁请求
			f.logger.Info("fsmNode recieve MsgArbitResponse")
		}
	}
	return nil
}
