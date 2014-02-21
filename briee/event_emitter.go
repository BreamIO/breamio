package briee

type EventEmitter interface {
	Publish(chid string, v interface{}) interface{}
	Subscribe(chid string, v interface{}) interface{}
	Run() // Runs the emitter
}

func NewEventEmitter() EventEmitter {
	return NewLocalEventEmitter()
}
