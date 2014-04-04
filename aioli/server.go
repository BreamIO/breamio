package aioli

import (
	"code.google.com/p/go.net/websocket"
	"log"
	"net"
	"net/http"
	"os"
	"path"
)

const (
	//Access server port for JSON encoded events over normal TCP connections.
	tcpJSONaddr = ":4041"
	
	//Access server port for JSON encoded events over WebSockets.
	wsJSONaddr  = ":8080"
)

// A Server is something that can listen.
// The intended usage is Event Access Servers, like WSServer and TCPServer.
type Server interface {
	Listen()
	//Close() Future update (ETA: 2037)
}

// Server is websocket server using the default decoder
type WSServer struct {
	manager IOManager
	logger  *log.Logger
}

// Creates a new WebSocket Event Access Server.
func NewWSServer(ioman IOManager, l *log.Logger) *WSServer {
	return &WSServer{
		manager: ioman,
		logger:  l,
	}
}

// Serve static files and listen for incoming websocket messages
func (s *WSServer) Listen() {
	pwd, err := os.Getwd()
	if err != nil {
		s.logger.Printf("Failed to get current working directory: %s\n", err)
		return
	}
	wsHandler := websocket.Handler(s.handler)
	fileHandler := http.FileServer(http.Dir(path.Join(pwd, "static")))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Handle websocket requests separately, but still serve static files
		if r.Header.Get("Upgrade") == "websocket" && r.Header.Get("Connection") == "Upgrade" {
			wsHandler.ServeHTTP(w, r)
		} else {
			fileHandler.ServeHTTP(w, r)
		}
	})
	s.logger.Printf("Listening on %s.", wsJSONaddr)
	err = http.ListenAndServe(wsJSONaddr, nil)
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

// This describes a Access Server over TCP.
type TCPServer struct {
	manager IOManager
	logger  *log.Logger
}

// Creates a new TCP Event Access Server.
func NewTCPServer(ioman IOManager, l *log.Logger) *TCPServer {
	return &TCPServer{ioman, l}
}

// Listen starts the TCP server, listening for incoming connections.
//
// When a connection is established, 
// it starts reading packages from it, handling them as it goes.
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
