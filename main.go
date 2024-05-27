package main

import (
	"fmt"
	"log"
	"net"
)

type Server struct {
	listenAddr string
	ln         net.Listener
	quitch     chan struct{}
	msgch      chan Message
}

type Message struct {
	from    string
	payload []byte
}

func main() {
	server := NewServer(":3000")

	go func() {
		for msg := range server.msgch {
			fmt.Printf("Received message from connection (%s): %s\n", msg.from, string(msg.payload))
		}
	}()

	log.Fatal(server.Start())
}

func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr: listenAddr,
		quitch:     make(chan struct{}),
		msgch:      make(chan Message, 10),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddr)

	if err != nil {
		return err
	}

	defer ln.Close()
	s.ln = ln
	go s.acceptLoop()
	<-s.quitch
	close(s.msgch)
	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept()

		if err != nil {
			fmt.Println("accept error:", err)
			continue
		}

		fmt.Println("New Connection: ", conn.RemoteAddr())
		go s.readLoop(conn)
	}
}

func (s *Server) readLoop(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 2048)

	for {
		n, err := conn.Read(buf)

		if err != nil {
			fmt.Println("read error:", err)
			return
		}

		s.msgch <- Message{
			from:    conn.RemoteAddr().String(),
			payload: buf[:n],
		}

		conn.Write([]byte("message received"))
		msg := buf[:n]
		fmt.Println(string(msg))
	}
}
