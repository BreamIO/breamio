package aioli

import (
	"code.google.com/p/go.net/websocket"
	"net/http"
)

// Server is websocket server using the default decoder
type Server struct {
	manager IOManager
}

func NewServer(ioman IOManager) *Server {
	return &Server{
		manager: ioman,
	}
}

// Listen and Serve for incomming message on the websocket.
func (s *Server) Listen() {
	http.Handle("/", websocket.Handler(s.handler))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

// handler is called for every established connection and will send data to the manager
func (s *Server) handler(ws *websocket.Conn) {
	dec := NewDecoder(ws)
	go s.manager.Listen(dec)
}
