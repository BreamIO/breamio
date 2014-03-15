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
	"strconv"
	"sync"
)

// The interface of a BreamIO logic.
// It allows you get access to the primary EventEmitter it listens to.
// In order for it to listen to anything, the ListenAndServe method must first be called
type Logic interface {
	RootEmitter() briee.EventEmitter
	ListenAndServe()
	io.Closer
}

// Creates a instance of the current Logic implementation.
func New(eef func() briee.EventEmitter, io aioli.IOManager) Logic {
	return newBL(eef, io)
}

type handlerFunc func(Spec) error


// First actual implementation
// Allows creation of trackers and statistics modules using the "new" event.
type breamLogic struct {
	root briee.EventEmitter
	ioman aioli.IOManager
	logger *log.Logger
	closer chan struct{}
	wg sync.WaitGroup
	onNewTrackerEvent handlerFunc
	eventEmitterConstructor func() briee.EventEmitter
}

func newBL(eef func() briee.EventEmitter, io aioli.IOManager) *breamLogic {
	logic := new(breamLogic)
	logic.logger = log.New(os.Stdout, "[Beenleigh] ", log.LstdFlags)
	logic.closer = make(chan struct{})
	logic.eventEmitterConstructor = eef
	
	//Create the first event emitter
	logic.root = eef()
	
	if io != nil {
		//Hook it up to the io manager
		logic.ioman = io
		logic.ioman.AddEE(logic.root, 256)
	}
	
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

func (bl *breamLogic) ListenAndServe() {
	defer bl.root.Close()
	//Subscribe to events
	newEvents := bl.root.Subscribe("new", Spec{}).(<-chan Spec)
	shutdownEvents := bl.root.Subscribe("shutdown", struct{}{}).(<-chan struct{})
	
	go bl.ioman.Run()
	
	//Set up servers.
	ts := aioli.NewTCPServer(bl.ioman, log.New(os.Stdout, "[TCPServer] ", log.LstdFlags))
	ws := aioli.NewWSServer(bl.ioman, log.New(os.Stdout, "[WSServer] ", log.LstdFlags))
	go ts.Listen()
	go ws.Listen()
	
	for {
		select {
			case event := <- newEvents:
				switch event.Type {
				case "tracker":
					if err := bl.onNewTrackerEvent(event); err != nil {
						bl.root.Dispatch("error:new:tracker", err)
					}
				case "statistics":
					/*if err := bl.onNewStatisticsEvent(event); err != nil {
						bl.root.Dispatch("error:new:tracker", err)
					}*/
				}
			case <- shutdownEvents:
				bl.logger.Println("Recieved shutdown event.")
				return
			case <- bl.closer:
				//bl.logger.Println("Time to close the shop!")
				return
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
	
	bl.wg.Add(1)
	ee := bl.eventEmitterConstructor()
	go func() {
		ee.Wait()
		bl.wg.Done()
	}()
	
	bl.ioman.AddEE(ee, event.Emitter)
	go tracker.Link(ee)
	
	//Should be moved to separate type handler
	go ancient.ListenAndServe(ee, byte(event.Emitter), ":303" + strconv.Itoa(event.Emitter))
	
	bl.logger.Printf("Created a new tracker with uri %s on EE %d.\n", event.Data, event.Emitter)
	return nil
}

func (bl *breamLogic) Close() error {
	close(bl.closer)
	bl.wg.Wait()
	return nil
}

// A specification for creation of new objects.
// Type should be a type available for creation by the logic implementation.
// Data is a context sensitive string, which syntax depends on the type.
// Emitter is a integer, identifying the emitter number to link the new object to. 
type Spec struct {
	Type string
	Data string
	Emitter int
}