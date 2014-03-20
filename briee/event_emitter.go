package briee

import (
	"reflect"
)

// EventEmitter interface contains methods for publishing, subscribring and managing events.
type EventEmitter interface {
	Publish(eventID string, v interface{}) chan<- interface{}
	Subscribe(eventID string, v interface{}) <-chan interface{}
	Unsubscribe(eventID string, <-chan ch interface{}) error
	TypeOf(eventID string) (reflect.Type, error)
	Dispatch(eventID string, v interface{})
	Close() error
	Wait()
}

// New creates a new instance of the default implementation LocalEventEmitter
func New() EventEmitter {
	return newLocalEventEmitter()
}
