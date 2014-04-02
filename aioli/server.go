package aioli

import (
	"code.google.com/p/go.net/websocket"
	"log"
	"net"
	"net/http"
)

const (
	tcpJSONaddr = ":4041"
	wsJSONaddr  = ":8080"
)

type Server interface {
	Listen()
	//Close() Future update (ETA: 2037)
}

// Server is websocket server using the default decoder
type WSServer struct {
	manager IOManager
	logger  *log.Logger
}

func NewWSServer(ioman IOManager, l *log.Logger) *WSServer {
	return &WSServer{
		manager: ioman,
		logger:  l,
	}
}

// Listen and Serve for incoming message on the websocket.
func (s *WSServer) Listen() {
	http.Handle("/", websocket.Handler(s.handler))
	s.logger.Printf("Listening on %s.", wsJSONaddr)
	err := http.ListenAndServe(wsJSONaddr, nil)
	if err != nil {
		s.logger.Printf("Failed to listen on TCP address %s: %s\n", tcpJSONaddr, err)
		return
	}
}

// handler is called for every established connection and will send data to the manager
func (s *WSServer) handler(ws *websocket.Conn) {
	codec := NewCodec(ws)
	go s.manager.Listen(codec, s.logger)
}

type TCPServer struct {
	manager IOManager
	logger  *log.Logger
}

func NewTCPServer(ioman IOManager, l *log.Logger) *TCPServer {
	return &TCPServer{ioman, l}
}

func (t *TCPServer) Listen() {
	ln, err := net.Listen("tcp", tcpJSONaddr)
	if err != nil {
		t.logger.Printf("Failed to listen on %s: %s\n", tcpJSONaddr, err)
		return
	}

	t.logger.Printf("Listening on %s.", tcpJSONaddr)
	defer ln.Close()
	for {
		//Check for closing
		in, err := ln.Accept()
		if err != nil {
			t.logger.Printf("Failed to accept connection on TCP address %s: %s\n", tcpJSONaddr, err)
			return
		}

		codec := NewCodec(in)
		go t.manager.Listen(codec, t.logger)
	}
}
