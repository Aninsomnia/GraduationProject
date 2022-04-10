package fsm

import (
	pb "GraduationProjection/fsm/fsmpb"
	"context"
)

type FSM interface {
	Tick()
	Step(ctx context.Context, msg pb.Message) error
	Ready() <-chan Ready
	Advance()
	Stop()
}

type fsm struct {
	recvc    chan pb.Message
	readyc   chan Ready
	advancec chan struct{}
	tickc    chan struct{}
	done     chan struct{}
	stop     chan struct{}

	fc *fsmCore
}
type Ready struct {
	Messages []pb.Message
}

func StartFsm(c *Config) FSM {
	fc := newfsmCore(c)
	f := &fsm{
		recvc:    make(chan pb.Message),
		readyc:   make(chan Ready),
		advancec: make(chan struct{}),
		tickc:    make(chan struct{}, 128),
		done:     make(chan struct{}),
		stop:     make(chan struct{}),
		fc:       fc,
	}
	go f.run()
	return f
}
func (f *fsm) run() {
	var readyc chan Ready
	var advancec chan struct{}
	var rd Ready

	for {
		// 此代码逻辑确保了当一个Ready发送给上层后，将会一直等待上层调用Advance，否则不会
		// 发送接下来的Ready
		if advancec != nil {
			readyc = nil
		} else if f.hasReady() {
			rd = f.newReady()
			readyc = f.readyc
		}
		select {
		case m := <-f.recvc:
			f.fc.Step(m)
		case <-f.tickc:
			// 收到时钟信号
			f.fc.tick()
		case readyc <- rd:
			f.fc.msgs = nil
			advancec = f.advancec
		case <-advancec:
			// TODO: 需要调用f.fc.advance(rd)吗
			rd = Ready{}
			advancec = nil
		case <-f.stop:
			close(f.done)
			return
		}
	}
}

// 由上层调用，进行一次时钟触发
func (f *fsm) Tick() {
	select {
	case f.tickc <- struct{}{}:
	case <-f.done:
	default:
		f.fc.logger.Warningf("A tick missed to fire. Node blocks too long!")
	}
}
func (f *fsm) Step(ctx context.Context, msg pb.Message) error {
	return nil
}

// 为调用者返回fsm的readyc
func (f *fsm) Ready() <-chan Ready {
	return f.readyc
}

func (f *fsm) Advance() {
	select {
	case f.advancec <- struct{}{}:
	case <-f.done:
	}
}
func (f *fsm) Stop() {
	select {
	// 还未准备好停止，所以先向f.stop通道触发停止信号
	case f.stop <- struct{}{}:
	case <-f.done:
		return
	}
	// Block until the stop has been acknowledged by run()
	<-f.done
}

func (f *fsm) hasReady() bool {
	return len(f.fc.msgs) > 0
}
func (f *fsm) newReady() Ready {
	rd := Ready{
		Messages: f.fc.msgs,
	}
	return rd
}
