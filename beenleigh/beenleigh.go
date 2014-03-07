package beenleigh

import (
	"github.com/maxnordlund/breamio/aioli"
	"github.com/maxnordlund/breamio/briee"
	"github.com/maxnordlund/breamio/gorgonzola"
	"log"
	"os"
)

var root briee.EventEmitter
var ioman aioli.IOManager

func RootEventEmitter() briee.EventEmitter {
	return root
}

func MainIOManager() aioli.IOManager {
	return ioman
}

func init() {
	//Create the first event emitter
	root = briee.NewLocalEventEmitter()
	
	//Hook it up to the io manager
	ioman = aioli.NewBasicIOManager()
	ioman.AddEE(root, 256)
}

var logger = log.New(os.Stdout, "[Beenleigh]", log.Ldate)

func ListenAndServe() {
	//Subscribe to events
	newTrackerEvents := root.Subscribe("new:tracker", TrackerSpec{}).(<-chan TrackerSpec)
	//newStatsEvents := root.Subscribe("new:statistics", StatsSpec{}).(<-chan StatsSpec)
	shutdownEvents := root.Subscribe("shutdown", TrackerSpec{}).(<-chan TrackerSpec)
	
	go ioman.Run()
	
	for {
		select {
			case event := <- newTrackerEvents:
				logger.Println("Recieved new:tracker event.")
				ee :=  briee.NewLocalEventEmitter()
				ioman.AddEE(ee, event.Number)
				tracker, err := gorgonzola.GetDriver(event.Type).CreateFromId(event.Id)
				if err != nil {
					logger.Printf("Could not create new tracker with type %s and id %s: %s", event.Type, event.Id, err)
					continue
				}
				tracker.Connect()
				go gorgonzola.Link(ee, tracker)
				logger.Printf("Created a new %s tracker with id %s on EE %d.\n", 
					event.Type, event.Id, event.Number)
			case <- shutdownEvents:
				logger.Println("Recieved shutdown event.")
				break
		}
	}
	logger.Println("Over and out!")
}

type TrackerSpec struct {
	Type string
	Id string
	Number int
}