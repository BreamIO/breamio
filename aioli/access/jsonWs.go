package access

import (
	"code.google.com/p/go.net/websocket"
	"github.com/maxnordlund/breamio/webber"

	// "net/http"

	"github.com/maxnordlund/breamio/aioli"
	"github.com/maxnordlund/breamio/beenleigh"
)

func init() {
	Register("WS(JSON)", &WSServer{})
}

// Server is websocket server using the default decoder
type WSServer struct {
	manager aioli.IOManager
	logger  module.Logger
}

// Serve static files and listen for incoming websocket messages
func (s *WSServer) Listen(ioman aioli.IOManager, logger module.Logger) {
	s.manager = ioman
	s.logger = logger
	webber.Instance().HandleWebSocket("/api/json", s.handler)

	s.logger.Println("Websocket server is up and running.")
}

// handler is called for every established connection and will send data to the manager
func (s *WSServer) handler(ws *websocket.Conn) {
	codec := aioli.NewCodec(ws)
	s.manager.Listen(codec, s.logger)
}
