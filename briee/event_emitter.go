package briee

import (
	"io"
	"reflect"
)

// Subscriber interface contains methods for subscribing and unsubscribing events.
type Subscriber interface {
	Subscribe(eventID string, v interface{}) interface{}
	Unsubscribe(eventID string, ch interface{}) error
}

// Publisher interface contains methods for publishing events.
type Publisher interface {
	Publish(eventID string, v interface{}) interface{}
}

// PublishSubscriber contains the Subscriber and Publisher interface.
type PublishSubscriber interface {
	Publisher
	Subscriber
}

// Dispatcher interface wraps the Dispatch method for sending a single event.
type Dispatcher interface {
	Dispatch(eventID string, v interface{})
}

// EventEmitter interface contains interfaces and methods for publishing, subscribing and managing events.
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
