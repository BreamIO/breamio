package access

import (
	"code.google.com/p/go.net/websocket"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/maxnordlund/breamio/aioli"
	//"github.com/maxnordlund/breamio/beenleigh"
)

const (
	//Access server port for JSON encoded events over WebSockets.
	wsJSONaddr = ":8080"
)

func init() {
	Register("WS(JSON)", &WSServer{})
}

// Server is websocket server using the default decoder
type WSServer struct {
	manager aioli.IOManager
	logger  *log.Logger
}

// Serve static files and listen for incoming websocket messages
func (s *WSServer) Listen(ioman aioli.IOManager, logger *log.Logger) {
	s.manager = ioman
	s.logger = logger

	pwd, err := os.Getwd()
	if err != nil {
		logger.Printf("Failed to get current working directory: %s\n", err)
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
	logger.Printf("Listening on %s.", wsJSONaddr)
	err = http.ListenAndServe(wsJSONaddr, nil)
	if err != nil {
		logger.Printf("Failed to listen on TCP address %s: %s\n", tcpJSONaddr, err)
		return
	}
}

// handler is called for every established connection and will send data to the manager
func (s *WSServer) handler(ws *websocket.Conn) {
	codec := aioli.NewCodec(ws)
	go s.manager.Listen(codec, s.logger)
}
