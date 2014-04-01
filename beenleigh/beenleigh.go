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

var constructers []Constructer

// Allows a module to register a constructor to be called during startup.
// The system also allows for destructors through the Close() error method.
// This is typically used to register global events and similar.
func Register(c Constructer) {
	constructers = append(constructers, c)
}

// The interface of a BreamIO logic.
// It allows you get access to the primary EventEmitter it listens to.
// In order for it to listen to anything, the ListenAndServe method must first be called
type Logic interface {
	RootEmitter() briee.EventEmitter
	CreateEmitter(id int) briee.EventEmitter
	aioli.EmitterLookuper
	ListenAndServe(aioli.IOManager)
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

func (bl *breamLogic) ListenAndServe(ioman aioli.IOManager) {
	defer bl.root.Close()

	//Subscribe to events
	for _, c := range constructers {
		go c.Init(bl)
		defer c.Close()
	}

	shutdownEvents := bl.root.Subscribe("shutdown", struct{}{}).(<-chan struct{})

	go ioman.Run()

	//Set up servers.
	ts := aioli.NewTCPServer(ioman, log.New(os.Stdout, "[TCPServer] ", log.LstdFlags))
	ws := aioli.NewWSServer(ioman, log.New(os.Stdout, "[WSServer] ", log.LstdFlags))
	go ts.Listen()
	go ws.Listen()

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
	close(bl.closer)
	bl.wg.Wait()
	return nil
}

// Creates a new emitter on the specified id if no such emitter exists.
// Regardless of pre-existence status, the emitter of that id is returned.
func (bl *breamLogic) CreateEmitter(id int) briee.EventEmitter {
	if _, ok := bl.emitters[id]; !ok {
		bl.wg.Add(1)
		bl.emitters[id] = bl.eventEmitterConstructor()
		go func() {
			bl.emitters[id].Wait()
			bl.wg.Done()
		}()

	}
	return bl.emitters[id]
}

func (bl *breamLogic) EmitterLookup(id int) (briee.EventEmitter, error) {
	if v, ok := bl.emitters[id]; ok {
		return v, nil
	}
	return nil, errors.New("No emitter with that id.")
}

type Constructer interface {
	Init(Logic)
	io.Closer
}

// A specification for creation of new objects.
// Type should be a type available for creation by the logic implementation.
// Data is a context sensitive string, which syntax depends on the type.
// Emitter is a integer, identifying the emitter number to link the new object to.
type Spec struct {
	Emitter int
	Data    string
}
