package briee

import (
	"errors"
	"reflect"
)

// Event is the internal representation of a registered event.
//
// It contains information about the publishing and subscribing
// channels, the channel element type and the number of publishers.
type Event struct {
	ElemType      reflect.Type                    // Element type
	PublSend      reflect.Value                   // Write only channel
	PublRecv      reflect.Value                   // Read only channel
	Subscribers   map[reflect.Value]reflect.Value // Map of mappings of SendCh to RecvCh
	NumPublishers int                             // Number of publishers
}

// LocalEventEmitter implements EventEmitter
type LocalEventEmitter struct {
	eventMap map[string]*Event
	closed   bool
	done     chan bool
	running  chan bool
}

// makeSendRecv returns two channels of provided channel element type.
// The two channels are write-only and read-only respectively.
func makeSendRecv(vtype reflect.Type) (chvSend, chvRecv reflect.Value) {
	// Get channel type
	chtype := reflect.ChanOf(reflect.BothDir, vtype)

	// Create the directed channel types
	chtypeSend := reflect.ChanOf(reflect.SendDir, vtype)
	chtypeRecv := reflect.ChanOf(reflect.RecvDir, vtype)

	// Make a two-way channel
	chv := reflect.MakeChan(chtype, 0)

	// Make a write-only channel
	chvSend = chv.Convert(chtypeSend)

	// Make a read-only channel
	chvRecv = chv.Convert(chtypeRecv)

	return
}

// Publish returns a write-only channel with element type equal to the underlying type of the provided interface.
//
// Data sent on the returned channel will be broadcasted to all subscribers of this event.
// An explicit type assertion of the returned channel is required if used in a non-reflective context.
// Will panic if called with an already registred event string identifier of unmatching types. Will also panic if type is not a struct or a pointer to a struct.
//
// Example use:
//
//		sendChan := ee.Publish("event string identifier", MyStruct{}).(chan<- MyStruct)
// 		sendChan <- MyStruct{...}
func (ee *LocalEventEmitter) Publish(eventID string, v interface{}) interface{} {
	// Get the type of v
	vtype := reflect.TypeOf(v)

	// TODO Refactor if performance is an issue. The channels/slice does not need to be constructed in all cases.
	chvSend, chvRecv := makeSendRecv(vtype)
	mapv := make(map[reflect.Value]reflect.Value)

	event, ok := ee.eventMap[eventID]

	if ok {
		if event.ElemType != vtype {
			panic("<Publish> Tried to publish on a existing event with different element types")
		}

		// Check if the event was created by a publisher
		if event.NumPublishers > 0 {
			event.NumPublishers += 1
			return event.PublSend.Interface()
		}
	} else {
		// Create a new event and store in map
		ee.eventMap[eventID] = &Event{
			ElemType: vtype,
			PublSend: reflect.Value{}, // Not assigned yet
			PublRecv: reflect.Value{}, // Not assigned yet
			//Subscribers:   slicev,
			Subscribers:   mapv,
			NumPublishers: 0,
		}

		event = ee.eventMap[eventID]
	}

	// The event exists at this point with the publisher channels missing
	ee.eventMap[eventID].PublSend = chvSend
	ee.eventMap[eventID].PublRecv = chvRecv
	ee.eventMap[eventID].NumPublishers += 1

	return chvSend.Interface()
}

func (ee *LocalEventEmitter) Dispatch(eventID string, value interface{}) {
	event := ee.eventMap[eventID]
	if event == nil {
		return
	}

	for _, sendCh := range event.Subscribers {
		sendCh.Send(reflect.ValueOf(value))
	}
}

// Subscribe returns a write-only channel with element type equal to the underlying type of the provided interface.
//
// An explicit type assertion of the returned channel is required if used in a non-reflective context.
// Will panic if called with an already registred event string identifier of unmatching types. Will also panic if type is not a struct or a pointer to a struct.
//
// Example use:
//		var recvData MyStruct
//		recvChan := ee.Subscribe("event string identifier", MyStruct{}).(<-chan MyStruct)
//		recvData = (<-recvChan)
func (ee *LocalEventEmitter) Subscribe(eventID string, v interface{}) interface{} {
	// Subscribe returns a read-only channel of element type of v

	// get the type of v
	vtype := reflect.TypeOf(v)

	// Make directed channels
	chvSend, chvRecv := makeSendRecv(vtype)

	event, ok := ee.eventMap[eventID]
	// Check if element is present
	if ok {
		if event.ElemType != vtype {
			panic("<Subscribe> Tried to subscribe on a existing event with different element type")
		}

	} else {
		// There are no events with this identifier, creates one
		//slicev := makeSlice(chvSend.Type())
		//mapv := makeMap(chvRecv.Type(), chvSend.Type())
		mapv := make(map[reflect.Value]reflect.Value)

		// Create a new event and store in map
		ee.eventMap[eventID] = &Event{
			ElemType: vtype,
			PublSend: reflect.Value{}, // Not assigned yet
			PublRecv: reflect.Value{}, // Not assigned yet
			//Subscribers:   slicev,
			Subscribers:   mapv,
			NumPublishers: 0,
		}

		event = ee.eventMap[eventID]
	}

	event.Subscribers[chvRecv] = chvSend

	// Return read only channel
	return chvRecv.Interface()

}

// Unsubscribe
func (ee *LocalEventEmitter) Unsubscribe(eventID string, ch interface{}) error {
	if event, ok := ee.eventMap[eventID]; ok {

		for recvCh, sendCh := range event.Subscribers {
			if recvCh == reflect.ValueOf(ch) {
				sendCh.Close()
				delete(event.Subscribers, recvCh)
				return nil
			}
		}

		return errors.New("Can not find subscriber")
	} else {
		// event not in event map
		return errors.New("Can not unsubscribe event which is not registered")
	}
}

// NewLocalEventEmitter is the constructor of LocalEventEmitter
//
// The emitter is open once constructed, IsClosed method call will return false.
func newLocalEventEmitter() *LocalEventEmitter {
	ee := &LocalEventEmitter{
		eventMap: make(map[string]*Event),
		closed:   true,
		done:     make(chan bool),
		running:  make(chan bool),
	}
	go ee.run()
	<-ee.running // Make sure the emitter is running before returning
	return ee
}

// Run method listens on publishing channels and broadcasts data to subscribers.
//
// This method is blocking until Close method is called.
func (ee *LocalEventEmitter) run() {
	ee.closed = false
	ee.running <- true

	for !ee.IsClosed() {

		// TODO Use a priority queue instead of linear range over the map?
		for _, event := range ee.eventMap {

			chv := event.PublRecv

			if !chv.IsValid() { // Checks that the channels isn't the zero value
				continue
			}

			recv, ok := chv.TryRecv()
			if ok {
				for _, sendCh := range event.Subscribers {
					sendCh.Send(recv)
				}
			}
		} // end for
	} // end if
	ee.done <- true
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

// Close will close all open channels to subscribers and terminate the Run method.
//
// Returns an error if emitter is already closed.
func (ee *LocalEventEmitter) Close() error {
	if ee.IsClosed() {
		return errors.New("Can not close already closed event emitter")
	}
	// Close the subscriber channels
	for _, event := range ee.eventMap {

		for recvCh, sendCh := range event.Subscribers {
			sendCh.Close()
			delete(event.Subscribers, recvCh)
		}

	}

	// Clear the event map
	for k := range ee.eventMap {
		delete(ee.eventMap, k)
	}

	ee.closed = true
	return nil
}

// IsClosed returns true if emitter is currently closed.
//
// Call Run method to open.
func (ee *LocalEventEmitter) IsClosed() bool {
	return ee.closed
}

// Wait will block until the emitter has been successfully closed
func (ee *LocalEventEmitter) Wait() {
	<-ee.done
}
