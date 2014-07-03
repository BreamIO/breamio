/*
Business Logic module

Beenleigh handles all business logic of BreamIO Eriver
It exposes a interface for interacting with this logic
but not the actual implementation.
*/
package beenleigh

import (
	"errors"
	"github.com/maxnordlund/breamio/aioli"
	"github.com/maxnordlund/breamio/briee"
	"io"
	"log"
	"os"
	"sync"
)

var runners []RunCloser

// Allows a module to register a constructor to be called during startup.
// The system also allows for destructors through the Close() error method.
// This is typically used to register global events and similar.
func Register(c RunCloser) {
	runners = append(runners, c)
}

// The interface of a BreamIO logic.
// It allows you get access to the primary EventEmitter it listens to.
// In order for it to listen to anything, the ListenAndServe method must first be called
type Logic interface {
	RootEmitter() briee.EventEmitter
	CreateEmitter(id int) briee.EventEmitter
	aioli.EmitterLookuper
	ListenAndServe()
	io.Closer
}

// Creates a instance of the current Logic implementation.
func New(eef func() briee.EventEmitter) Logic {
	return newBL(eef)
}

// First actual implementation
// Allows creation of trackers and statistics modules using the "new" event.
type breamLogic struct {
	root                    briee.EventEmitter
	logger                  *log.Logger
	closer                  chan struct{}
	wg                      sync.WaitGroup
	eventEmitterConstructor func() briee.EventEmitter
	emitters                map[int]briee.EventEmitter
	lock                    sync.RWMutex
}

func newBL(eef func() briee.EventEmitter) *breamLogic {
	logic := new(breamLogic)
	logic.logger = log.New(os.Stdout, "[Beenleigh] ", log.LstdFlags)
	logic.closer = make(chan struct{})
	logic.emitters = make(map[int]briee.EventEmitter)
	logic.eventEmitterConstructor = eef

	//Create the first event emitter
	logic.root = eef()
	logic.emitters[256] = logic.root

	return logic
}

func (bl *breamLogic) RootEmitter() briee.EventEmitter {
	return bl.root
}

func (bl *breamLogic) ListenAndServe() {
	defer bl.root.Close()

	//Subscribe to events
	for _, c := range runners {
		go c.Run(bl)
		defer c.Close()
	}

	shutdownEvents := bl.root.Subscribe("shutdown", struct{}{}).(<-chan struct{})

	//Set up servers.
	//ts := aioli.NewTCPServer(ioman, log.New(os.Stdout, "[TCPServer] ", log.LstdFlags))
	//ws := aioli.NewWSServer(ioman, log.New(os.Stdout, "[WSServer] ", log.LstdFlags))
	//go ts.Listen()
	//go ws.Listen()

	for {
		select {
		case <-shutdownEvents:
			bl.logger.Println("Recieved shutdown event.")
			bl.Close()
			return
		case <-bl.closer:
			//bl.logger.Println("Time to close the shop!")
			return
		}
	}
}

func (bl *breamLogic) Close() error {
	defer bl.lock.Unlock()
	bl.lock.Lock()
	close(bl.closer)
	bl.wg.Wait()
	return nil
}

// Creates a new emitter on the specified id if no such emitter exists.
// Regardless of pre-existence status, the emitter of that id is returned.
func (bl *breamLogic) CreateEmitter(id int) briee.EventEmitter {
	defer bl.lock.Unlock()
	bl.lock.Lock()
	emitter, ok := bl.emitters[id]
	if !ok {
		emitter = bl.eventEmitterConstructor()
		bl.wg.Add(1)
		bl.emitters[id] = emitter
		go func(emitter briee.EventEmitter) {
			emitter.Wait()
			bl.wg.Done()
		}(emitter)

	}
	return emitter
}

func (bl *breamLogic) EmitterLookup(id int) (briee.EventEmitter, error) {
	defer bl.lock.RUnlock()
	bl.lock.RLock()
	if v, ok := bl.emitters[id]; ok {
		return v, nil
	}
	return nil, errors.New("No emitter with that id.")
}
