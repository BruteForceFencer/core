// Package server implements a server following its own custom protocol.  The
// protocol works as follows.
//
//     1. A client connects and sends an object in JSON format matching the
//     structure of the Request type.
//
//     2. The server returns a single character. "t" means that the request is
//     valid.  "f" means that the request is invalid (either an attack or an
//     error).
//
//     3. The client disconnects or goes again from step 1.
package controlserver

import (
	"log"
	"net"
	"os"
)

// Server is a server that interprets requests according to the protocol.
type Server struct {
	HandleFunc func(*Request) bool
	listener   net.Listener
}

// Blocks and listens for requests.
func (s *Server) ListenAndServe(typ, addr string) error {
	// Remove any old socket.
	if typ == "unix" {
		os.Remove(addr)
	}

	// Start listening.
	var err error
	s.listener, err = net.Listen(typ, addr)
	if err != nil {
		return err
	}

	// Accept requests.
	go s.acceptRequests()
	return nil
}

func (s *Server) acceptRequests() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Println("connection error:", err)
			continue
		}

		// For performance, we launch every handler in its own goroutine.
		go func(conn net.Conn) {
			for {
				request, err := ReadRequest(conn)
				if err != nil {
					conn.Close()
					return
				}

				response := s.HandleFunc(request)
				if response {
					conn.Write([]byte("t"))
				} else {
					conn.Write([]byte("f"))
				}
			}
		}(conn)
	}
}

// Close stops the server.
func (s *Server) Close() {
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
		}
	}()
	s.listener.Close()
}