package access

import (
	"log"
	"net"

	"github.com/maxnordlund/breamio/aioli"
)

//Access server port for JSON encoded events over normal TCP connections.
const tcpJSONaddr = ":4041"

func init() {
	registerTCPJSON()
}

func registerTCPJSON() {
	Register("TCP(JSON)", TCPServer{})
}

type TCPServer struct{}

// Listen starts the TCP server, listening for incoming connections.
//
// When a connection is established,
// it starts reading packages from it, handling them as it goes.
func (t TCPServer) Listen(ioman aioli.IOManager, logger *log.Logger) {
	ln, err := net.Listen("tcp", tcpJSONaddr)
	if err != nil {
		logger.Printf("Failed to listen on %s: %s\n", tcpJSONaddr, err)
		return
	}

	logger.Printf("Listening on %s.", tcpJSONaddr)
	defer ln.Close()
	for {
		//Check for closing
		in, err := ln.Accept()
		if err != nil {
			logger.Printf("Failed to accept connection on TCP address %s: %s\n", tcpJSONaddr, err)
			return
		}

		codec := aioli.NewCodec(in)
		go ioman.Listen(codec, logger)
	}
}
