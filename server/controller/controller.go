package controller

type Controller struct {
	f        fsmNode
	stop     chan struct{}
	stopping chan struct{}
	done     chan struct{}
}

func NewController() (*Controller, error) {
	b, err := bootstrap()
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
