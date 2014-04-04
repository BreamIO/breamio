package aioli

import (
	"github.com/maxnordlund/breamio/briee"
	"log"
)

// ExtPkg is the struct used as the external protocol
type ExtPkg struct {
	Event string // Name of the event
	Subscribe bool // Should the handler setup a subscription channel for this event and client.
	ID    int    // Event Emitter identifier, 0 for broadcast
	Data  []byte // Encoded data of the underlying struct for the event.
}

// Defines something that can be used to retrieve EventEmitters.
type EmitterLookuper interface {
	EmitterLookup(int) (briee.EventEmitter, error)
}

// IOManager interface defines an I/O manager with external reader functionality.
type IOManager interface {
	Listen(codec EncodeDecoder, l *log.Logger)
	Run()
	Close() error
}

// New creates a new instance of the default implementation BasicIOManager
func New(lookuper EmitterLookuper) IOManager {
	return newBasicIOManager(lookuper)
}
