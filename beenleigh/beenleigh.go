package beenleigh

import (
	"github.com/maxnordlund/breamio/aioli"
	"github.com/maxnordlund/breamio/briee"
	"github.com/maxnordlund/breamio/gorgonzola"
	"log"
	"os"
	"io"
)

type Logic interface {
	RootEmitter() briee.EventEmitter
	MainIOManager() aioli.IOManager
	ListenAndServe()
	io.Closer
}

func New() Logic {
	return newBL()
}

type handlerFunc func(Spec) error

type breamLogic struct {
	root briee.EventEmitter
	ioman aioli.IOManager
	logger *log.Logger
	closer chan struct{}
	onNewTrackerEvent handlerFunc
}

// Constructor function for EventEmitters.
// Allows Dependency Injection for testing purposes. 
var newee = func() briee.EventEmitter {
	return briee.New()
}

// Constructor function for IOManagers.
// Allows Dependency Injection for testing purposes. 
var newio = func() aioli.IOManager {
	return aioli.New()
}

func newBL() *breamLogic {
	logic := new(breamLogic)
	logic.logger = log.New(os.Stdout, "[Beenleigh]", log.Ldate)
	logic.closer = make(chan struct{})
	
	//Create the first event emitter
	logic.root = newee()
	
	//Hook it up to the io manager
	logic.ioman = newio()
	logic.ioman.AddEE(logic.root, 256)
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

func (bl *breamLogic) MainIOManager() aioli.IOManager {
	return bl.ioman
}

func (bl *breamLogic) ListenAndServe() {
	//Subscribe to events
	newEvents := bl.root.Subscribe("new", Spec{}).(<-chan Spec)
	shutdownEvents := bl.root.Subscribe("shutdown", struct{}{}).(<-chan struct{})
	
	go bl.ioman.Run()
	
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
	ee :=  newee()
	bl.ioman.AddEE(ee, event.Emitter)
	tracker, err := gorgonzola.CreateFromURI(event.Data)
	if err != nil {
		bl.logger.Printf("Could not create new tracker with uri %s: %s", event.Data, err)
		return err
	}
	tracker.Connect()
	go tracker.Link(ee)
	bl.logger.Printf("Created a new tracker with uri %s on EE %d.\n", event.Data, event.Emitter)
	return nil
}

func (bl *breamLogic) Close() error {
	close(bl.closer)
	return nil
}

type Spec struct {
	Type string
	Data string
	Emitter int
}