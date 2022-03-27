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
	// 接收心跳判断是否失效的超时时间
	HeartbeatReceiveTick int
	logger               Logger
}

type fsmCore struct {
	ID     uint64
	peerID uint64

	heartbeatSendTimeout int
	heartbeatSendElapsed int

	heartbeatReceiveTimeout int
	heartbeatReceiveElapsed int

	state StateType
	msgs  []pb.Message

	step   stepFunc
	logger Logger
}

func newfsmCore(c *Config) *fsmCore {
	f := &fsmCore{
		ID:                      c.ID,
		peerID:                  c.peerID,
		heartbeatSendTimeout:    c.HeartbeatSendTick,
		heartbeatReceiveTimeout: c.HeartbeatReceiveTick,
		logger:                  c.logger,
	}
	f.becomeSingleRunning()
	f.logger.Infof("fsmCore %x was created", f.ID)
	return f
}
func (f *fsmCore) Step(m pb.Message) error {
	switch m.Type {
	case pb.MsgBeat:
		f.sendHeartbeat()

	// 对于pb.MsgHeartbeat类型
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
	// 在dualRunning状态下收到心跳，若心跳包中指明对方etcd已故障：
	// 1）状态将变为pending，2）发起仲裁请求
	case pb.MsgHeartbeat:
		f.heartbeatReceiveElapsed = 0
		if m.IsEtcdStoped {
			f.becomePending()
			// TODO:发起仲裁
			return nil
		}
	}

	return nil
}
func stepPending(f *fsmCore, m pb.Message) error {
	switch m.Type {
	// 在pending状态下收到心跳，若心跳包中指明对方etcd未故障：
	// 1）状态变为dualRunning，2）解除仲裁
	case pb.MsgHeartbeat:
		f.heartbeatReceiveElapsed = 0
		if !m.IsEtcdStoped {
			f.becomeDualRunning()
			// TODO:解除仲裁
		}
	}
	return nil
}
func stepSingleRunning(f *fsmCore, m pb.Message) error {
	switch m.Type {
	// 在singleRunning状态下收到心跳，若心跳包中指明对方etcd未故障：
	// 1）执行扩容，2）解除仲裁
	case pb.MsgHeartbeat:
		f.heartbeatReceiveElapsed = 0
		if !m.IsEtcdStoped {
			f.becomeDualRunning()
			// TODO:扩容
		}

	}
	return nil
}
func stepEtcdStoped(f *fsmCore, m pb.Message) error {
	switch m.Type {
	case pb.MsgHeartbeat:
		return nil

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
	f.heartbeatReceiveElapsed++

	// 判断是否该发送心跳
	if f.heartbeatSendElapsed >= f.heartbeatSendTimeout {
		f.heartbeatSendElapsed = 0
		if err := f.Step(pb.Message{Type: pb.MsgBeat}); err != nil {
			f.logger.Debugf("error occurred during checking sending heartbeat: %v", err)
		}
	}
	// 判断是否已经无法接收来自对方的心跳
	if f.heartbeatReceiveElapsed >= f.heartbeatReceiveTimeout {
		f.heartbeatReceiveElapsed = 0
		//TODO 剩余处理
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
