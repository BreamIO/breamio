/*
Business Logic module

Beenleigh handles all business logic of BreamIO Eriver
It exposes a interface for interacting with this logic 
but not the actual implementation.
*/
package beenleigh

import (
	"github.com/maxnordlund/breamio/aioli"
	ancient "github.com/maxnordlund/breamio/aioli/ancientPower"
	"github.com/maxnordlund/breamio/briee"
	"github.com/maxnordlund/breamio/gorgonzola"
	"log"
	"os"
	"io"
	"errors"
	"strconv"
	"sync"
)

// The interface of a BreamIO logic.
// It allows you get access to the primary EventEmitter it listens to.
// In order for it to listen to anything, the ListenAndServe method must first be called
type Logic interface {
	RootEmitter() briee.EventEmitter
	aioli.EmitterLookuper
	ListenAndServe(aioli.IOManager)
	io.Closer
}

// Creates a instance of the current Logic implementation.
func New(eef func() briee.EventEmitter) Logic {
	return newBL(eef)
}

type handlerFunc func(Spec) error


// First actual implementation
// Allows creation of trackers and statistics modules using the "new" event.
type breamLogic struct {
	root briee.EventEmitter
	logger *log.Logger
	closer chan struct{}
	wg sync.WaitGroup
	onNewTrackerEvent handlerFunc
	eventEmitterConstructor func() briee.EventEmitter
	emitters map[int]briee.EventEmitter
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

	logic.onNewTrackerEvent = func() handlerFunc {
		return func(spec Spec) error {
			return onNewTrackerEvent(logic, spec)
		}
	}()
	
	return logic
}

func (bl *breamLogic) RootEmitter() briee.EventEmitter {
	return bl.root
}

func (bl *breamLogic) ListenAndServe(ioman aioli.IOManager) {
	defer bl.root.Close()
	//Subscribe to events
	
	shutdownEvents := bl.root.Subscribe("shutdown", struct{}{}).(<-chan struct{})
	
	go ioman.Run()
	
	//Set up servers.
	ts := aioli.NewTCPServer(ioman, log.New(os.Stdout, "[TCPServer] ", log.LstdFlags))
	ws := aioli.NewWSServer(ioman, log.New(os.Stdout, "[WSServer] ", log.LstdFlags))
	go ts.Listen()
	go ws.Listen()
	
	go bl.handle("new:tracker", bl.onNewTrackerEvent)
	go bl.handle("new:ancientpower", bl.onNewAncientEvent)
	
	for {
		select {
			case <- shutdownEvents:
				bl.logger.Println("Recieved shutdown event.")
				bl.Close()
				return
			case <- bl.closer:
				//bl.logger.Println("Time to close the shop!")
				return
		}
	}
}

func (bl *breamLogic) handle(eventId string, f handlerFunc) {
	bl.wg.Add(1)
	defer bl.wg.Done()
	
	events := bl.root.Subscribe(eventId, Spec{}).(<-chan Spec)
	for {
		select {
			case <- bl.closer: return
			case spec := <- events: 
				if err := f(spec); err != nil {
					bl.root.Dispatch("beenleigh:error", err)
				}
		}
	}
}

func onNewTrackerEvent(bl *breamLogic, event Spec) error {
	bl.logger.Println("Recieved new:tracker event.")
	
	tracker, err := gorgonzola.CreateFromURI(event.Data)
	if err != nil {
		bl.logger.Printf("Could not create new tracker with uri %s: %s", event.Data, err)
		return err
	}
	err = tracker.Connect()
	if err != nil {
		bl.logger.Println("Unable to connect to tracker:", err)
		return err
	}
	
	ee := getEmitter(event.Emitter)
	go tracker.Link(bl.emitters[event.Emitter])
	
	bl.logger.Printf("Created a new tracker with uri %s on EE %d.\n", event.Data, event.Emitter)
	return nil
}

func onNewAncientEvent(bl *breamLogic, event Spec) error {
	ee := bl.getEmitter(event.Emitter)
	go ancient.ListenAndServe(ee , byte(event.Emitter), add)
	//addr := ":303" + strconv.Itoa(event.Emitter)
	addr := event.Data
	bl.logger.Printf("Created a new AncientPower Server address %s on EE %d.\n", addr, event.Emitter)
	return nil
}

func (bl *breamLogic) Close() error {
	close(bl.closer)
	bl.wg.Wait()
	return nil
}

func (bl *breamLogic) getEmitter(id int) briee.EventEmitter {
	if _, ok := bl.emitters[event.Emitter]; !ok {
		bl.wg.Add(1)
		bl.emitters[event.Emitter] = bl.eventEmitterConstructor()
		go func() {
			bl.emitters[event.Emitter].Wait()
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

// A specification for creation of new objects.
// Type should be a type available for creation by the logic implementation.
// Data is a context sensitive string, which syntax depends on the type.
// Emitter is a integer, identifying the emitter number to link the new object to. 
type Spec struct {
	Emitter int
	Data string
}