package servermain

import (
	"GraduationProject/handle"
	"GraduationProject/node"
	"GraduationProject/serverhttp"
)

type Server struct {
	transport serverhttp.Transporter
	handle    handle.Handler
	localetcd node.Localetcd
	stop      chan struct{}
	stopping  chan struct{}
	done      chan struct{}
}

func NewServer() (s *Server, err error) {
	return
}

func (s *Server) StartServer() {

}
