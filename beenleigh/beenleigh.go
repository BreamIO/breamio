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

type handlerFunc func(interface{}) (done bool)

type breamLogic struct {
	root briee.EventEmitter
	ioman aioli.IOManager
	logger *log.Logger
	closer chan struct{}
}

func New() Logic {
	logic := new(breamLogic)
	logic.logger = log.New(os.Stdout, "[Beenleigh]", log.Ldate)
	logic.closer = make(chan struct{})
	
	//Create the first event emitter
	logic.root = briee.NewLocalEventEmitter()
	
	//Hook it up to the io manager
	logic.ioman = aioli.NewBasicIOManager()
	logic.ioman.AddEE(logic.root, 256)
	
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
				bl.onNewTrackerEvent(event)
			case <- shutdownEvents:
				bl.logger.Println("Recieved shutdown event.")
				return
			case <- bl.closer:
				//bl.logger.Println("Time to close the shop!")
				return
		}
	}
}

func (bl *breamLogic) onNewTrackerEvent(event TrackerSpec) bool {
	bl.logger.Println("Recieved new:tracker event.")
	ee :=  briee.NewLocalEventEmitter()
	bl.ioman.AddEE(ee, event.Number)
	tracker, err := gorgonzola.GetDriver(event.Type).CreateFromId(event.Id)
	if err != nil {
		bl.logger.Printf("Could not create new tracker with type %s and id %s: %s", event.Type, event.Id, err)
		return false
	}
	tracker.Connect()
	go gorgonzola.Link(ee, tracker)
	bl.logger.Printf("Created a new %s tracker with id %s on EE %d.\n", 
		event.Type, event.Id, event.Number)
	return false
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