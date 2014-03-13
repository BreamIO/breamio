package aioli

import (
	"github.com/maxnordlund/breamio/briee"
	"log"
)

// ExtPkg is the struct used as the external protocol
type ExtPkg struct {
	Event string // Name of the event
	ID    int    // Event Emitter identifier, 0 for broadcast
	Data  []byte // Encoded data of the underlying struct for the event.
}

// IOManager interface defines an I/O manager with external reader functionality
type IOManager interface {
	//Listen(r io.Reader)
	Listen(dec Decoder, l *log.Logger)
	Run()
	AddEE(ee briee.EventEmitter, id int) error
	RemoveEE(id int) error
	Close() error
}

// New creates a new instance of the default implementation BasicIOManager
func New() IOManager {
	return newBasicIOManager()
}
