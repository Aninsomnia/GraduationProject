package fsm

import (
	pb "GraduationProjection/fsm/fsmpb"
)

type StateType uint64

const (
	dualRunning StateType = iota
	pending
	singleRunning
	etcdStoped
)

type Config struct {
	ID     uint64
	peerID uint64
	// 发送心跳的周期超时时间
	HeartbeatSendTick int
	logger            Logger
}

type fsmCore struct {
	ID     uint64
	peerID uint64

	heartbeatSendTimeout int
	heartbeatSendElapsed int

	state StateType
	msgs  []pb.Message

	step   stepFunc
	logger Logger
}

func newfsmCore(c *Config) *fsmCore {
	f := &fsmCore{
		ID:                   c.ID,
		peerID:               c.peerID,
		heartbeatSendTimeout: c.HeartbeatSendTick,
		logger:               c.logger,
	}
	f.becomeSingleRunning()
	f.logger.Infof("fsmCore %x was created", f.ID)
	return f
}
func (f *fsmCore) Step(m pb.Message) error {
	switch m.Type {
	case pb.MsgBeat:
		f.sendHeartbeat()

	// 对于pb.MsgHeartbeat、MsgEtcdPend、MsgEtcdStop、MsgEtcdReCovery类型
	default:
		if err := f.step(f, m); err != nil {
			return err
		}

	}
	return nil
}

type stepFunc func(f *fsmCore, m pb.Message) error

func stepDualRunning(f *fsmCore, m pb.Message) error {
	switch m.Type {
	case pb.MsgEtcdPend:
		f.becomePending()
	case pb.MsgEtcdStop:
		f.becomeEtcdStoped()

	}
	return nil
}
func stepPending(f *fsmCore, m pb.Message) error {
	switch m.Type {
	case pb.MsgEtcdStop:
		f.becomeEtcdStoped()
	case pb.MsgEtcdReCovery:
		f.becomeDualRunning()
	}
	return nil
}
func stepSingleRunning(f *fsmCore, m pb.Message) error {
	switch m.Type {
	//该状态下不可能收到MsgEtcdPend
	//该状态下不可能收到MsgEtcdReCovery
	case pb.MsgEtcdStop:
		f.becomeEtcdStoped()

	}
	return nil
}
func stepEtcdStoped(f *fsmCore, m pb.Message) error {
	switch m.Type {
	//该状态下不可能收到MsgEtcdPend
	//该状态下不可能收到MsgEtcdStop
	//该状态下不可能收到MsgEtcdReCovery

	}
	return nil
}

func (f *fsmCore) send(m pb.Message) {
	f.msgs = append(f.msgs, m)
}

func (f *fsmCore) sendHeartbeat() {
	isEtcdStoped := (f.state == etcdStoped)
	m := pb.Message{
		To:           f.peerID,
		Type:         pb.MsgHeartbeat,
		IsEtcdStoped: isEtcdStoped,
	}
	f.send(m)
}

func (f *fsmCore) requestArbitration() {

}

func (f *fsmCore) tick() {
	f.heartbeatSendElapsed++

	// 判断是否该发送心跳
	if f.heartbeatSendElapsed >= f.heartbeatSendTimeout {
		f.heartbeatSendElapsed = 0
		if err := f.Step(pb.Message{Type: pb.MsgBeat}); err != nil {
			f.logger.Debugf("error occurred during checking sending heartbeat: %v", err)
		}
	}
}
func (f *fsmCore) reset() {
	f.heartbeatSendElapsed = 0

}
func (f *fsmCore) becomeDualRunning() {
	f.reset()
	f.step = stepDualRunning
	f.state = dualRunning

}
func (f *fsmCore) becomePending() {
	f.reset()
	f.step = stepPending
	f.state = pending

}
func (f *fsmCore) becomeSingleRunning() {
	f.reset()
	f.step = stepSingleRunning
	f.state = singleRunning
}
func (f *fsmCore) becomeEtcdStoped() {
	f.reset()
	f.step = stepEtcdStoped
	f.state = etcdStoped
}
