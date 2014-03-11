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

type handlerFunc func(TrackerSpec) error

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
		return func(ts TrackerSpec) error {
			return onNewTrackerEvent(logic, ts)
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
	newTrackerEvents := bl.root.Subscribe("new:tracker", TrackerSpec{}).(<-chan TrackerSpec)
	//newStatsEvents := bl.root.Subscribe("new:statistics", StatsSpec{}).(<-chan StatsSpec)
	shutdownEvents := bl.root.Subscribe("shutdown", TrackerSpec{}).(<-chan TrackerSpec)
	
	go bl.ioman.Run()
	
	for {
		select {
			case event := <- newTrackerEvents:
				if err := bl.onNewTrackerEvent(event); err != nil {
					
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

func onNewTrackerEvent(bl *breamLogic, event TrackerSpec) error {
	bl.logger.Println("Recieved new:tracker event.")
	ee :=  newee()
	bl.ioman.AddEE(ee, event.Number)
	tracker, err := gorgonzola.GetDriver(event.Type).CreateFromId(event.Id)
	if err != nil {
		bl.logger.Printf("Could not create new tracker with type %s and id %s: %s", event.Type, event.Id, err)
		return err
	}
	tracker.Connect()
	go gorgonzola.Link(ee, tracker)
	bl.logger.Printf("Created a new %s tracker with id %s on EE %d.\n", 
		event.Type, event.Id, event.Number)
	return nil
}

func (bl *breamLogic) Close() error {
	close(bl.closer)
	return nil
}

type TrackerSpec struct {
	Type string
	Id string
	Number int
}