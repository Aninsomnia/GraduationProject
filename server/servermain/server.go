package servermain

import "GraduationProjection/server/controller"

type server struct {
	c    *controller.Controller
	errc chan error
}

// startServer将会创建一个server，并且运行它，然后再监听其是否stop
func startServer() (<-chan struct{}, <-chan error, error) {
	var err error
	s := &server{}

	s.c, err = controller.NewController()
	if err != nil {
		return nil, nil, err
	}

	s.c.Start()

	select {
	case <-s.c.StopNotify():
	}
	return s.c.StopNotify(), s.errc, nil

}
