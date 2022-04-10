package controller

import "go.uber.org/zap"

type Controller struct {
	f        fsmNode
	logger   *zap.Logger
	stop     chan struct{}
	stopping chan struct{}
	done     chan struct{}
}

func NewController(cfg *ControllerConfig) (*Controller, error) {
	b, err := bootstrap(cfg)
	if err != nil {
		return nil, err
	}
	return &Controller{
		f: *b.fsm.newFsmNode(),
	}, nil
}

func (c *Controller) Start() {

	c.done = make(chan struct{})
	c.stop = make(chan struct{})
	c.stopping = make(chan struct{}, 1)

	go c.run()
}
func (c *Controller) run() {
	// 启动fsmNode
	c.f.start()
}

func (c *Controller) StopNotify() <-chan struct{}     { return c.done }
func (c *Controller) StoppingNotify() <-chan struct{} { return c.stopping }
