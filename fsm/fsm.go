package fsm

import (
	pb "GraduationProjection/fsm/fsmpb"
)

type StateType uint64

const (
	dualActive StateType = iota
	dualInActive
	singleActive
	localDead
)

type Config struct {
	localID       uint64
	peerID        uint64
	HeartbeatTick int
	logger        Logger
}

type fsm struct {
	localID          uint64
	peerID           uint64
	heartbeatTimeout int
	heartbeatElapsed int
	state            StateType
	msgs             []pb.Message
	logger           Logger
}

func newfsm(c *Config) *fsm {
	f := &fsm{
		localID:          c.localID,
		peerID:           c.peerID,
		heartbeatTimeout: c.HeartbeatTick,
		logger:           c.logger,
	}
	f.becomeDualInActive()
	f.logger.Infof("fsm %x was created", f.localID)
	return f
}
func (f *fsm) Step(m pb.Message) error {

}

func (f *fsm) send() {

}
func (f *fsm) sendHeartbeat() {

}
func (f *fsm) tickHeartbeat() {
	f.heartbeatElapsed++
	if f.heartbeatElapsed >= f.heartbeatTimeout {
		f.heartbeatElapsed = 0
		if err := f.Step(pb.Message{Type: pb.MsgBeat, From: f.localID}); err != nil {
			f.logger.Debugf("error occurred during checking sending heartbeat: %v", err)
		}
	}
}
func (f *fsm) reset() {
	f.heartbeatElapsed = 0

}
func (f *fsm) becomeDualActive() {
	f.reset()
	f.state = dualActive

}
func (f *fsm) becomeDualInActive() {
	f.reset()
	f.state = dualInActive

}
func (f *fsm) becomeSingleActive() {
	f.reset()
	f.state = singleActive
}
func (f *fsm) becomeLocalDead() {
	f.reset()
	f.state = localDead
}
