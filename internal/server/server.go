package server

import (
	"fmt"
	"log"
	"net"
)

type Config struct {
	Port string
}

type Server struct {
	config *Config
}

func NewServer(c *Config) *Server {
	return &Server{
		config: c,
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s", s.config.Port))
	if err != nil {
		return fmt.Errorf("could not start listener", err)
	}

	return s.process(listener)
}

func (s *Server) process(l net.Listener) error {
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		go s.handleConn(conn)
	}

	// return nil
}

func (s *Server) handleConn(conn net.Conn) {
	go s.readLoop(conn)
	go s.writeLoop(conn)
}


/*
Notes:
* Server should know what has been sent to which device last (this information can be saved in a database)
* Timestamp of the last update should be set at the start of the synchronization between client and server
* Client should keep the last time it synchronized with the server
	* Ask the server for all metadata files changed since this timestamp
1) Send all metadata files to find the discrepancies
2) 
*/
func (s *Server) readLoop(conn net.Conn) {
	for {
		b := make([]byte, 10)
		read, err := conn.Read(b)
		if err != nil {
			log.Println("could not read from connection: %w", err)
			break
		}
	}
}

func (s *Server) writeLoop(conn net.Conn) {
	for {
		// b := make([]byte, 10)
		// read, err := conn.Read(b)
		// if err != nil {
		// 	log.Println("could not read from connection: %w", err)
		// 	break
		// }
	}
}
