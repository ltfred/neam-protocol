package server

import (
	"errors"
	"fmt"
	"github.com/ltfred/neam-protocol/message"
	"github.com/sirupsen/logrus"
	"net"
)

type Config struct {
	Host string
	Port int64
}

type Server struct {
	cfg      *Config
	listener net.Listener
}

func NewServer(cfg *Config) (*Server, error) {
	if cfg.Host == "" || cfg.Port == 0 {
		return nil, errors.New("server not configured")
	}
	s := &Server{cfg: cfg}
	return s, nil
}

func (s *Server) Run() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port))
	if err != nil {
		logrus.Error(err)

		return
	}
	s.listener = listener

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			logrus.Error(err)

			return
		}

		go s.process(conn)
	}
}

func (s *Server) process(conn net.Conn) {
	_ = message.NewMessage(conn)

}
