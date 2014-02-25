package briee

import (
	"reflect"
)

type EventEmitter interface {
	Publish(chid string, v interface{}) interface{}
	Subscribe(chid string, v interface{}) interface{}
	TypeOf(eventID string) (reflect.Type, error)
	Run() // Runs the emitter
}

func NewEventEmitter() EventEmitter {
	return NewLocalEventEmitter()
}
