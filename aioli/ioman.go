package aioli

import (
	"github.com/maxnordlund/breamio/briee"
)

type ExtPkg struct {
	// Struct used as external protocol
	Event string // Name of the event
	ID    int    // EE identifier, 0 for broadcast
	Data  []byte // Data of the underlying struct for the event.
}

type IOManager interface {
	// Methods of the IO manager
	Listen(recvCh <-chan ExtPkg) // Run, ListenAndServe?
	AddEE(ee *briee.EventEmitter, id int) error
	RemoveEE(id int) error
}

func NewIOManager() IOManager {
	return NewBasicIOManager()
}
