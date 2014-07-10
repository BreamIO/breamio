package mozzarella

import	(
	"log"
)

type EventAnalyser struct {
	stopCh	chan struct{}
	closeCh chan struct{}
	resetCh chan struct{}
}

func NewEventAnalyser() *EventAnalyser {
	return &EventAnalyser{
			stopCh: make(chan {}),
			closeCh: make(chan {}),
			resetCh: make(chan {}),
		}
}

func (e *EventAnalyser) Run() {
	log := log.New(os.Stderr, "[ EventAnalyser ] ", log.LstdFlags)
}

// Start analysing events in the program
func (e *EventAnalyser) start(emitters []briee.EventEmitter) {
	log.Println("Starting analysis of EEs") //TODO might wanna print out wich EEs
	for _, ee := range emitters {
		ee.Subscribe("new:regionStats", new(regionStats.Config)).(<-chan *regionStats.Config)
		defer ee.Unsubscribe("new:regionStats", newChan)
		//TODO subs and unsubs to ALL messages. =)
	}
}

// Stop analysing events in the program
func (e *EventAnalyser) stop() {

}

// Clear all data currently collected in the program
func (e *EventAnalyser) reset() {

}
