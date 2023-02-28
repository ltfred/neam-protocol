package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type Server struct {
	cfg      *Config
	listener net.Listener
	clients  map[uint64]*clientConn
	rwlock   sync.RWMutex

	deviceConnectionID map[string]uint64
}

func NewServer(cfg *Config) (s *Server, err error) {
	s = &Server{
		cfg:                cfg,
		clients:            make(map[uint64]*clientConn),
		deviceConnectionID: make(map[string]uint64),
	}

	if s.cfg.Host == "" || s.cfg.Port == 0 {
		return nil, errors.New("server not configured")
	}

	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	if s.listener, err = net.Listen("tcp", addr); err != nil {
		return nil, err
	}

	logrus.Infof("server is running NEAM protocol, addr: %s", addr)

	return s, nil
}

func (s *Server) Run() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok {
				if opErr.Err.Error() == "use of closed network connection" {
					return
				}
			}

			logrus.Fatalf("accept failed, err: %s", err)

			return
		}

		clientConn, err := s.newConn(conn)
		if err != nil {
			logrus.Errorf("create conn err: %s", err)

			continue
		}

		go s.onConn(clientConn)
	}
}

func (s *Server) Close() {
	s.rwlock.Lock()
	defer s.rwlock.Unlock()

	if s.listener != nil {
		err := s.listener.Close()
		if err != nil {
			logrus.Errorf("Close listener faild, err: %s", err)
		}

		s.listener = nil
	}
}

func (s *Server) GracefulDown(ctx context.Context, done chan struct{}) {
	logrus.Infof("[server] graceful shutdown")

	count := s.ConnectionCount()

	for i := 0; count > 0; i++ {
		s.kickIdleConnection()

		count = s.ConnectionCount()
		if count == 0 {
			break
		}

		ticker := time.After(time.Second)
		select {
		case <-ctx.Done():
			return
		case <-ticker:
		}
	}
}

func (s *Server) kickIdleConnection() {
	var conns []*clientConn

	s.rwlock.RLock()
	for _, cc := range s.clients {
		if cc.ShutdownOrNotify() {
			conns = append(conns, cc)
		}
	}
	s.rwlock.RUnlock()

	for _, cc := range conns {
		err := cc.Close()
		if err != nil {
			logrus.Errorf("close connection err: %s", err)
		}
	}
}

func (s *Server) ConnectionCount() int {
	s.rwlock.RLock()
	cnt := len(s.clients)
	s.rwlock.RUnlock()

	return cnt
}

func (s *Server) newConn(conn net.Conn) (*clientConn, error) {
	cc, err := newClientConn(s)
	if err != nil {
		return nil, err
	}

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		if err := tcpConn.SetKeepAlive(s.cfg.TCPKeepAlive); err != nil {
			logrus.Errorf("failed to set tcp keep alive option, err: %s", err)
		}

		if err := tcpConn.SetNoDelay(s.cfg.TCPNoDelay); err != nil {
			logrus.Errorf("failed to set tcp no delay option, err %s", err)
		}
	}

	cc.setConn(conn)

	return cc, nil
}

func (s *Server) onConn(conn *clientConn) {
	logrus.Debugf("new connection, remoteAddr: %s, connection id: %d",
		conn.bufReadConn.RemoteAddr().String(),
		conn.connectionID)

	defer func() {
		logrus.Debug("connection closed")
	}()

	s.rwlock.Lock()
	s.clients[conn.connectionID] = conn
	s.rwlock.Unlock()

	conn.Run()
}
