package transport

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"
)

type Server struct {
	started chan struct{}
}

func NewServer() *Server {
	return &Server{
		started: make(chan struct{}),
	}
}

func (s *Server) Run(ctx context.Context) error {
	l, err := net.Listen("tcp", ":4332")
	if err != nil {
		return err
	}
	close(s.started)

	return s.doLoop(ctx, l)
}

func (s *Server) WaitStart() {
	<-s.started
}

func (s *Server) doLoop(ctx context.Context, listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		childCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		go s.handleConn(childCtx, conn)
	}
}

func (s *Server) handleConn(ctx context.Context, conn net.Conn) {
	go func() {
		for {
			data := make([]byte, 100)
			conn.SetReadDeadline(time.Now().Add(time.Second * 5))
			n, err := conn.Read(data)
			if err == context.DeadlineExceeded {
				log.Println("read timed out...")
			} else if err != nil {
				log.Println("read errored", err)
			}

			fmt.Println(n, string(data[:n]))
		}
	}()
}
