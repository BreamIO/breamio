package briee

import (
	"reflect"
)

// EventEmitter interface.
type EventEmitter interface {
	Publish(chid string, v interface{}) interface{}
	Subscribe(chid string, v interface{}) interface{}
	TypeOf(eventID string) (reflect.Type, error)
	Close() error
	Run()
}

// NewEventEmitter creates and returns a new LocalEventEmitter
func New() EventEmitter {
	return newLocalEventEmitter()
}
