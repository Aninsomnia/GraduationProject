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
	fc := &fsmCore{
		heartbeatSendTimeout:    c.HeartbeatSendTick,
		reqArbitSendTimeout:     c.ReqArbitSendTick,
		heartbeatReceiveTimeout: c.HeartbeatReceiveTick,
	}

	fc.heartbeatSendElapsed = 0
	fc.heartbeatReceiveElapsed = 0
	fc.reqArbitSendElapsed = 0
	fc.logger = fsmLogger
	fc.becomeFaultPending()
	fc.logger.Infof("fsmCore was created")
	return fc
}
func (fc *fsmCore) Step(m pb.Message) error {
	switch m.Type {
	default:
		if err := fc.step(fc, m); err != nil {
			return err
		}
	}
	return nil
}

type stepFunc func(fc *fsmCore, m pb.Message) error

func stepDualRunning(fc *fsmCore, m pb.Message) error {
	switch m.Type {
	case pb.MsgMemberAddResp:
		return nil
	case pb.MsgHeartbeat:
		// dualRunning状态下收到了对方的异常心跳
		if m.State != pb.DualRunning {
			fc.becomeFaultPending()
		}
	}
	return nil
}
func stepFaultPending(fc *fsmCore, m pb.Message) error {
	switch m.Type {

	}
	return nil
}
func stepSingleRunning(fc *fsmCore, m pb.Message) error {
	switch m.Type {
	case pb.MsgMemberAddResp:
		// recoveryPending状态下收到了MemberAddResp，说明之前发送的MemberAdd被对方接收
		fc.becomeRecoveryPending()
		// TODO:执行扩容
	}
	return nil
}
func stepRecoveryPending(fc *fsmCore, m pb.Message) error {
	switch m.Type {

	}
	return nil
}

func stepEtcdStoped(fc *fsmCore, m pb.Message) error {
	switch m.Type {
	case pb.MsgMemberAdd:
		// etcdStoped状态下收到了对方的MsgMemberAdd，说明对方已经尝试扩容集群
		fc.becomeRecoveryPending()
		// TODO:执行扩容
	}
	return nil
}

func (fc *fsmCore) send(m pb.Message) {
	fc.msgs = append(fc.msgs, m)
}

func (fc *fsmCore) requestArbitration() {
	m := pb.Message{
		Type: pb.MsgArbitRequest,
	}
	fc.send(m)
}

func (fc *fsmCore) releaseArbitration() {
	m := pb.Message{
		Type: pb.MsgArbitRelease,
	}
	fc.send(m)
}

func (fc *fsmCore) tickDualRunning() {
	fc.heartbeatSendElapsed++
	fc.heartbeatReceiveElapsed++

	// 判断是否该发送心跳
	if fc.heartbeatSendElapsed >= fc.heartbeatSendTimeout {
		fc.heartbeatSendElapsed = 0
		fc.send(pb.Message{Type: pb.MsgHeartbeat, State: pb.DualRunning})
	}

	// 判断来自对方的心跳是否超时
	if fc.heartbeatReceiveElapsed >= fc.heartbeatReceiveTimeout {
		fc.heartbeatReceiveElapsed = 0
		// 状态转变
		fc.becomeFaultPending()
	}
}
func (fc *fsmCore) tickFaultPending() {
	fc.heartbeatSendElapsed++
	fc.reqArbitSendElapsed++

	// 判断是否该发送心跳
	if fc.heartbeatSendElapsed >= fc.heartbeatSendTimeout {
		fc.heartbeatSendElapsed = 0
		fc.send(pb.Message{Type: pb.MsgHeartbeat, State: pb.FaultPending})
	}
	// 判断是否该发送仲裁请求
	if fc.reqArbitSendElapsed >= fc.reqArbitSendTimeout {
		fc.reqArbitSendElapsed = 0
		fc.send(pb.Message{Type: pb.MsgArbitRequest})
	}
}
func (fc *fsmCore) tickSingleRunning() {
	fc.heartbeatSendElapsed++

	// 判断是否该发送心跳
	if fc.heartbeatSendElapsed >= fc.heartbeatSendTimeout {
		fc.heartbeatSendElapsed = 0
		fc.send(pb.Message{Type: pb.MsgMemberAdd, State: pb.SingleRunning})
	}
}
func (fc *fsmCore) tickRecoveryPending() {
	fc.heartbeatSendElapsed++

	// 判断是否该发送心跳，并且通知member add
	if fc.heartbeatSendElapsed >= fc.heartbeatSendTimeout {
		fc.heartbeatSendElapsed = 0
		// 如果是新加入的节点
		if !fc.isHasArbitrationLock {
			fc.send(pb.Message{Type: pb.MsgMemberAddResp, State: pb.RecoveryPending})
		}

	}
}
func (fc *fsmCore) tickEtcdStoped() {
	fc.heartbeatSendElapsed++

	// 判断是否该发送心跳
	if fc.heartbeatSendElapsed >= fc.heartbeatSendTimeout {
		fc.heartbeatSendElapsed = 0
		fc.send(pb.Message{Type: pb.MsgHeartbeat, State: pb.EtcdStoped})
	}
}

func (fc *fsmCore) becomeDualRunning() {

	fc.state = dualRunning
	fc.step = stepDualRunning
	fc.tick = fc.tickDualRunning
}
func (fc *fsmCore) becomeFaultPending() {
	fc.state = faultPending
	fc.step = stepFaultPending
	fc.tick = fc.tickFaultPending
}
func (fc *fsmCore) becomeRecoveryPending() {

	fc.state = recoveryPending
	fc.step = stepRecoveryPending
	fc.tick = fc.tickRecoveryPending

}
func (fc *fsmCore) becomeSingleRunning() {

	fc.state = singleRunning
	fc.step = stepSingleRunning
	fc.tick = fc.tickSingleRunning
}
func (fc *fsmCore) becomeEtcdStoped() {

	fc.state = etcdStoped
	fc.step = stepEtcdStoped
	fc.tick = fc.tickEtcdStoped
}

func (fc *fsmCore) advance(rd Ready) {

}
