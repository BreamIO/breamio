package briee

import (
	"reflect"
)

// EventEmitter interface.
type EventEmitter interface {
	Publish(eventID string, v interface{}) interface{}
	Subscribe(eventID string, v interface{}) interface{}
	TypeOf(eventID string) (reflect.Type, error)
	Dispatch(eventID string, v interface{})
	Close() error
	Run()
}

// New creates a new instance of the default implementation LocalEventEmitter
func New() EventEmitter {
	return newLocalEventEmitter()
}
