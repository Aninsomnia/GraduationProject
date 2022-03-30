package fsm

import (
	pb "GraduationProjection/fsm/fsmpb"
)

type StateType uint64

const (
	dualRunning StateType = iota
	faultPending
	recoveryPending
	singleRunning
	etcdStoped
)

type Config struct {
	// 发送心跳的周期超时时间
	HeartbeatSendTick int

	HeartbeatReceiveTick int
	// 仲裁请求的发送周期
	ReqArbitSendTick int
	logger           Logger
}

type fsmCore struct {
	heartbeatSendTimeout int
	heartbeatSendElapsed int

	heartbeatReceiveTimeout int
	heartbeatReceiveElapsed int

	reqArbitSendTimeout int
	reqArbitSendElapsed int

	isHasArbitrationLock bool
	state                StateType
	msgs                 []pb.Message

	step   stepFunc
	tick   func()
	logger Logger
}

func newfsmCore(c *Config) *fsmCore {
	f := &fsmCore{
		heartbeatSendTimeout:    c.HeartbeatSendTick,
		reqArbitSendTimeout:     c.ReqArbitSendTick,
		heartbeatReceiveTimeout: c.HeartbeatReceiveTick,
		logger:                  c.logger,
	}

	f.heartbeatSendElapsed = 0
	f.heartbeatReceiveElapsed = 0
	f.reqArbitSendElapsed = 0

	f.becomeDualRunning()
	f.logger.Infof("fsmCore %x was created")
	return f
}
func (f *fsmCore) Step(m pb.Message) error {
	switch m.Type {
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
	case pb.MsgMemberAddResp:
		return nil
	case pb.MsgHeartbeat:
		// dualRunning状态下收到了对方的异常心跳
		if m.State != pb.DualRunning {
			f.becomeFaultPending()
		}
	}
	return nil
}
func stepFaultPending(f *fsmCore, m pb.Message) error {
	switch m.Type {

	}
	return nil
}
func stepSingleRunning(f *fsmCore, m pb.Message) error {
	switch m.Type {
	case pb.MsgMemberAddResp:
		// recoveryPending状态下收到了MemberAddResp，说明之前发送的MemberAdd被对方接收
		f.becomeRecoveryPending()
		// TODO:执行扩容
	}
	return nil
}
func stepRecoveryPending(f *fsmCore, m pb.Message) error {
	switch m.Type {

	}
	return nil
}

func stepEtcdStoped(f *fsmCore, m pb.Message) error {
	switch m.Type {
	case pb.MsgMemberAdd:
		// etcdStoped状态下收到了对方的MsgMemberAdd，说明对方已经尝试扩容集群
		f.becomeRecoveryPending()
		// TODO:执行扩容
	}
	return nil
}

func (f *fsmCore) send(m pb.Message) {
	f.msgs = append(f.msgs, m)
}

func (f *fsmCore) requestArbitration() {
	m := pb.Message{
		Type: pb.MsgReqArbit,
	}
	f.send(m)
}

func (f *fsmCore) releaseArbitration() {
	m := pb.Message{
		Type: pb.MsgReleaseArbit,
	}
	f.send(m)
}

func (f *fsmCore) tickDualRunning() {
	f.heartbeatSendElapsed++
	f.heartbeatReceiveElapsed++

	// 判断是否该发送心跳
	if f.heartbeatSendElapsed >= f.heartbeatSendTimeout {
		f.heartbeatSendElapsed = 0
		f.send(pb.Message{Type: pb.MsgHeartbeat, State: pb.DualRunning})
	}

	// 判断来自对方的心跳是否超时
	if f.heartbeatReceiveElapsed >= f.heartbeatReceiveTimeout {
		f.heartbeatReceiveElapsed = 0
		// 状态转变
		f.becomeFaultPending()
	}
}
func (f *fsmCore) tickFaultPending() {
	f.heartbeatSendElapsed++
	f.reqArbitSendElapsed++

	// 判断是否该发送心跳
	if f.heartbeatSendElapsed >= f.heartbeatSendTimeout {
		f.heartbeatSendElapsed = 0
		f.send(pb.Message{Type: pb.MsgHeartbeat, State: pb.FaultPending})
	}
	// 判断是否该发送仲裁请求
	if f.reqArbitSendElapsed >= f.reqArbitSendTimeout {
		f.reqArbitSendElapsed = 0
		f.send(pb.Message{Type: pb.MsgReqArbit})
	}
}
func (f *fsmCore) tickSingleRunning() {
	f.heartbeatSendElapsed++

	// 判断是否该发送心跳
	if f.heartbeatSendElapsed >= f.heartbeatSendTimeout {
		f.heartbeatSendElapsed = 0
		f.send(pb.Message{Type: pb.MsgMemberAdd, State: pb.SingleRunning})
	}
}
func (f *fsmCore) tickRecoveryPending() {
	f.heartbeatSendElapsed++

	// 判断是否该发送心跳，并且通知member add
	if f.heartbeatSendElapsed >= f.heartbeatSendTimeout {
		f.heartbeatSendElapsed = 0
		// 如果是新加入的节点
		if !f.isHasArbitrationLock {
			f.send(pb.Message{Type: pb.MsgMemberAddResp, State: pb.RecoveryPending})
		}

	}
}
func (f *fsmCore) tickEtcdStoped() {
	f.heartbeatSendElapsed++

	// 判断是否该发送心跳
	if f.heartbeatSendElapsed >= f.heartbeatSendTimeout {
		f.heartbeatSendElapsed = 0
		f.send(pb.Message{Type: pb.MsgHeartbeat, State: pb.EtcdStoped})
	}
}

func (f *fsmCore) becomeDualRunning() {

	f.state = dualRunning
	f.step = stepDualRunning
	f.tick = f.tickDualRunning
}
func (f *fsmCore) becomeFaultPending() {
	f.state = faultPending
	f.step = stepFaultPending
	f.tick = f.tickFaultPending
}
func (f *fsmCore) becomeRecoveryPending() {

	f.state = recoveryPending
	f.step = stepRecoveryPending
	f.tick = f.tickRecoveryPending

}
func (f *fsmCore) becomeSingleRunning() {

	f.state = singleRunning
	f.step = stepSingleRunning
	f.tick = f.tickSingleRunning
}
func (f *fsmCore) becomeEtcdStoped() {

	f.state = etcdStoped
	f.step = stepEtcdStoped
	f.tick = f.tickEtcdStoped
}
