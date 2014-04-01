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
	dec := NewDecoder(ws)
	go s.manager.Listen(dec, s.logger)
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

		dec := NewDecoder(in)
		go t.manager.Listen(dec, t.logger)
	}
}
