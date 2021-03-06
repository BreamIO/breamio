package access

import (
	"encoding/gob"
	"io"
	"net"

	"github.com/maxnordlund/breamio/moduler"
	"github.com/maxnordlund/breamio/remote"
)

//Access server port for JSON encoded events over normal TCP connections.
const tcpJSONaddr = ":4041"

func init() {
	registerTCPJSON()
	//registerTCPGOB()
}

func registerTCPJSON() {
	Register("TCP(JSON)", TCPServer{func(conn io.ReadWriteCloser) remote.EncodeDecoder {
		return remote.NewCodec(conn)
	}})
}

func registerTCPGOB() {
	Register("TCP(GOB)", TCPServer{func(conn io.ReadWriteCloser) remote.EncodeDecoder {
		return remote.Codec{gob.NewEncoder(conn), gob.NewDecoder(conn)}
	}})
}

type TCPServer struct {
	codecConstructor func(io.ReadWriteCloser) remote.EncodeDecoder
}

// Listen starts the TCP server, listening for incoming connections.
//
// When a connection is established,
// it starts reading packages from it, handling them as it goes.
func (t TCPServer) Listen(ioman remote.IOManager, logger moduler.Logger) {
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
		logger.Printf("Connection recieved from %s.", in.RemoteAddr())
		go ioman.Listen(t.codecConstructor(in), logger)
	}
}
