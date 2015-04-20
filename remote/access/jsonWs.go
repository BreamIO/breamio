package access

import (
	"code.google.com/p/go.net/websocket"
	"github.com/maxnordlund/breamio/webserver"

	// "net/http"

	"github.com/maxnordlund/breamio/moduler"
	"github.com/maxnordlund/breamio/remote"
)

func init() {
	Register("WS(JSON)", &WSServer{})
}

// Server is websocket server using the default decoder
type WSServer struct {
	manager remote.IOManager
	logger  moduler.Logger
}

// Serve static files and listen for incoming websocket messages
func (s *WSServer) Listen(ioman remote.IOManager, logger moduler.Logger) {
	s.manager = ioman
	s.logger = logger
	webserver.Instance().HandleWebSocket("/api/json", s.handler)

	s.logger.Println("Websocket server is up and running.")
}

// handler is called for every established connection and will send data to the manager
func (s *WSServer) handler(ws *websocket.Conn) {
	codec := remote.NewCodec(ws)
	s.manager.Listen(codec, s.logger)
}
