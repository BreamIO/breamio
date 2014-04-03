package briee

import (
	"errors"
	"reflect"
)

// Event is the internal representation of an event on the event emitter. 
//
// An event contains information about the underlying type sent on publishing and subscribing channels but also the subscribing channels them selfs.
// The publishing channels are administrated by stand-alone goroutines and the calling publisher.
type Event struct {
	ElemType reflect.Type // Underlying element type
	DataChan    reflect.Value // Channel used for the internal data
	Subscribers reflect.Value // Slice of write-only channels to subscribers
	CanSubscribe bool // Boolean indicating if the overhead subscriber goroutine is running or not.
	SubscriberMap map[reflect.Value]reflect.Value // Subscriber binding to enable unsubscription.
}

// newEvent is the constructor of an event.
//
// The constructor takes a reflect.Type as the only parameter and corresponds to the underlying type to be sent on the generated channels.
func newEvent(elemtype reflect.Type) *Event {
	return &Event{
		ElemType: elemtype,
		DataChan:    makeChan(elemtype),
		Subscribers: reflect.Value{},
		CanSubscribe: false,
		SubscriberMap: make(map[reflect.Value]reflect.Value),
	}
}

// LocalEventEmitter implements the EventEmitter interface.
type LocalEventEmitter struct {
	eventMap map[string]*Event
	open     bool
	done     chan struct{}
}

// newLocalEventEmitter is the constructor for the LocalEventEmitter.
//
// The event emitter is open when constructed.
func newLocalEventEmitter() *LocalEventEmitter {
	return &LocalEventEmitter{
		eventMap: make(map[string]*Event),
		open:     true,
		done:     make(chan struct{}),
	}
}


// Publish returns a write-only channel with element type equal to the underlying type of the provided interface.
//
// Data sent on the returned channel will be broadcasted to all subscribers of this event.
// An explicit type assertion of the returned channel is required if used in a non-reflective context.
// Will panic if called with an already registred event string identifier of unmatching types.
//
// Example use:
//
//		sendChan := ee.Publish("event string identifier", MyStruct{}).(chan<- MyStruct)
// 		sendChan <- MyStruct{...}
func (ee *LocalEventEmitter) Publish(eventID string, v interface{}) interface{} {
	event := ee.event(eventID, v)

	// Create the write and read channels
	sendChan, recvChan := makeDirChannels(v, 0)

	go func() {
		for {
			if data, okRecv := recvChan.Recv(); okRecv && ee.IsOpen() { // Recv is blocking
				event.DataChan.TrySend(data)
			} else {
				break
			}
		}
	}()

	return sendChan.Interface()
}

// Subscribe returns a read-only channel with element type equal to the underlying type of the provided interface.
//
// An explicit type assertion of the returned channel is required if used in a non-reflective context.
// Will panic if called with an already registred event string identifier of unmatching types. 
//
// Example use:
//		var recvData MyStruct
//		recvChan := ee.Subscribe("event string identifier", MyStruct{}).(<-chan MyStruct)
//		recvData = (<-recvChan)
func (ee *LocalEventEmitter) Subscribe(eventID string, v interface{}) interface{} {

	event := ee.event(eventID, v)

	// Create the write and read channels
	sendChan, recvChan := makeDirChannels(v, 256)

	event.SubscriberMap[recvChan] = sendChan

	if !event.CanSubscribe {

		event.Subscribers = makeSlice(reflect.TypeOf(v))
		event.CanSubscribe = true

		go func() {
			defer func(){
				for i := 0; i < event.Subscribers.Len(); i++ {
					ch := event.Subscribers.Index(i)
					ch.Close()
				}
			}()

			for {
				if data, ok := event.DataChan.Recv(); ok && ee.IsOpen(){
					for i := 0; i < event.Subscribers.Len(); i++ {
						if ee.IsOpen() {
							ch := event.Subscribers.Index(i)
							ch.TrySend(data)
						} else {
							return
						}
					}
				} else {
					return
				}
			}
		}()
	}

	event.Subscribers = reflect.Append(event.Subscribers, sendChan)
	return recvChan.Interface()
}


// Dispatch will perform a one-time publish to all listening subscribers.
//
// The underlying value of the provided interface will be sent.
// Will panic if the event string identifier has a different registered type as the underlying type of the provided interface.
// The method call is not blocking.
// Example:
//		ee.Dispatch("event string identifier", MyStruct{...})
func (ee *LocalEventEmitter) Dispatch(eventID string, v interface{}) {
	if ee.IsOpen() {
		event := ee.event(eventID, v)
		event.DataChan.TrySend(reflect.ValueOf(v))
	}
}

// TypeOf returns the reflect.Type registered for the requested event string
// identifier. Will return nil and an error if requested event is not present.
//
// Example:
//		var rtype reflect.Type
//		rtype, err := ee.TypeOf("event string identifier")
//		if err != nil {
//			fmt.Println(err)
//		}
func (ee *LocalEventEmitter) TypeOf(eventID string) (reflect.Type, error) {
	if event, ok := ee.eventMap[eventID]; ok {
		return event.ElemType, nil
	} else {
		return nil, errors.New("No event with that identifier is registred")
	}
}

// Unsubscribe takes the event string identifier and the channel returned from an call to Subscribe and removes it from the registered event.
//
// Unsubscribe will return an error if the event string identifier is not existing. Will also return an error if provided channel is not registered on that event.
func (ee *LocalEventEmitter) Unsubscribe(eventID string, ch interface{}) error {
	// Check if event exisits
	if event, ok := ee.eventMap[eventID]; ok {
		recvChan := reflect.ValueOf(ch)
		sendChan, ook := event.SubscriberMap[recvChan]

		// Check if a mapping exisits
		if !ook {
			return errors.New("Can not find subscriber")
		}

		// Find the write channel and close it
		for i := 0; i < event.Subscribers.Len(); i++ {
			if event.Subscribers.Index(i).Interface() == sendChan.Interface() {
				sendChan.Close()
				delete(event.SubscriberMap, recvChan)
				return nil
			}
		}
	}

	// No event with that eventID exists
	return errors.New("Can not unsubscribe unregistered event")
}

// Close will close all subscribing channels of all registered events.
//
// Will return an error if Close is called on an already closed event emitter.
func (ee *LocalEventEmitter) Close() error {
	select {
	case <-ee.done:
		return errors.New("Emitter already closed")
	default:
		ee.open = false
		for _, event := range ee.eventMap {
			event.DataChan.Close()
		}
		close(ee.done)
	}

	return nil
}

// Wait is a blocking call and will wait until the Close method has been called.
func (ee *LocalEventEmitter) Wait() {
	// Wait for close to finish
	<-ee.done
}

// IsOpen returns true if the event emitter is open, else false.
//
// The emitter is open once constructed and closed when the Close method is called.
func (ee *LocalEventEmitter) IsOpen() bool {
	return ee.open
}

// event returns the registered Event in the Event map. Will create a new event if no event is present.
func (ee *LocalEventEmitter) event(eventID string, v interface{}) *Event {
	vtype := reflect.TypeOf(v)
	event, ok := ee.eventMap[eventID]
	if ok {
		// Event exisits, check that element types are consistent
		if event.ElemType != vtype {
			panic("Cannot send or receive data on an existing event with different element types")
		}
	} else {
		// No existing event, create one and store in map
		event = newEvent(vtype)
		ee.eventMap[eventID] = event
	}
	return event
}

// makeDirChannels returns two channels of provided channel element type.
//
// The two channels are write-only and read-only respectively.
func makeDirChannels(v interface{}, buffer int) (reflect.Value, reflect.Value) {
	vtype := reflect.TypeOf(v)

	// Get channel type
	chtype := reflect.ChanOf(reflect.BothDir, vtype)

	// Create the directed channel types
	chtypeSend := reflect.ChanOf(reflect.SendDir, vtype)
	chtypeRecv := reflect.ChanOf(reflect.RecvDir, vtype)

	// Make a two-way channel
	chv := reflect.MakeChan(chtype, buffer)

	// Make a write-only channel
	chvSend := chv.Convert(chtypeSend)

	// Make a read-only channel
	chvRecv := chv.Convert(chtypeRecv)

	return chvSend, chvRecv

}

// makeChan returns a reflect.Value of a channel of the provided reflect.Type as element type.
//
// Is created with a 256 buffer.
func makeChan(vtype reflect.Type) reflect.Value {
	chtype := reflect.ChanOf(reflect.BothDir, vtype)
	chv := reflect.MakeChan(chtype, 256)
	return chv
}

// makeSlice returns a reflect.Value of a slice of the provided reflect.Type as element type.
//
// Is created with an empty buffer.
func makeSlice(vtype reflect.Type) reflect.Value {
	chtype := reflect.ChanOf(reflect.SendDir, vtype) // Note SendDir
	slicetype := reflect.SliceOf(chtype)

	slicev := reflect.MakeSlice(slicetype, 0, 0)

	return slicev
}
