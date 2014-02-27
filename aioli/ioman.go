package aioli

import (
	"github.com/maxnordlund/breamio/briee"
	"io"
)

type ExtPkg struct {
	// Struct used as external protocol
	Event string // Name of the event
	ID    int    // EE identifier, 0 for broadcast
	Data  []byte // Data of the underlying struct for the event.
}

type IOManager interface {
	// Methods of the IO manager
	/* New layout
	Listen(r io.Reader), does the current Listen functionallity internaly in Run
	Run()
	*/
	//Listen(r io.Reader)
	//Run()
	//Listen(recvCh <-chan ExtPkg) // Run, ListenAndServe?
	Listen(r io.Reader)
	Run()
	AddEE(ee *briee.EventEmitter, id int) error
	RemoveEE(id int) error
}

func NewIOManager() IOManager {
	return NewBasicIOManager()
}

/*
// TODO
type Decoder interface {
	Decode()
}
*/
