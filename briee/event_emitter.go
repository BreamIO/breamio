package briee

import (
	"io"
	"reflect"
)

type Subscriber interface {
	Subscribe(eventID string, v interface{}) interface{}
	Unsubscribe(eventID string, ch interface{}) error
}

type Publisher interface {
	Publish(eventID string, v interface{}) interface{}
}

type PublishSubscriber interface {
	Publisher
	Subscriber
}

type Dispatcher interface {
	Dispatch(eventID string, v interface{})
}

// EventEmitter interface contains methods for publishing, subscribring and managing events.
type EventEmitter interface {
	PublishSubscriber
	Dispatcher
	TypeOf(eventID string) (reflect.Type, error)
	io.Closer
	Wait()
}

// New creates a new instance of the default implementation LocalEventEmitter
func New() EventEmitter {
	return newLocalEventEmitter()
}
