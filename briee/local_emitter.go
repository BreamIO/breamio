package briee

import (
	"errors"
	"log"
	"reflect"
)

// Event is the internal representation of a registered event.
//
// It contains information about the publishing and subscribing
// channels, the channel element type and the number of publishers.
type Event struct {
	ElemType      reflect.Type  // Element type
	PublSend      reflect.Value // Write only channel
	PublRecv      reflect.Value // Read only channel
	Subscribers   reflect.Value // Slice of write only channels
	NumPublishers int           // Number of publishers
}

// LocalEventEmitter implements EventEmitter
type LocalEventEmitter struct {
	eventMap map[string]*Event
	closed   bool
}

// makeSendRecv returns two channels of provided channel element type.
// The two channels are write-only and read-only respectively.
func makeSendRecv(vtype reflect.Type) (chvSend, chvRecv reflect.Value) {
	// Get channel type
	chtype := reflect.ChanOf(reflect.BothDir, vtype)

	// Create the directed channel types
	chtypeSend := reflect.ChanOf(reflect.SendDir, vtype)
	chtypeRecv := reflect.ChanOf(reflect.RecvDir, vtype)

	if !chtype.ConvertibleTo(chtypeSend) {
		log.Panic("<Publish> Cannot convert bi-directional channel to write-only\n")
	}
	if !chtype.ConvertibleTo(chtypeRecv) {
		log.Panic("<Publish> Cannot convert bi-directional channel to read-only\n")
	}

	// Make a two-way channel
	chv := reflect.MakeChan(chtype, 0)

	// Make a write-only channel
	chvSend = chv.Convert(chtypeSend)

	// Make a read-only channel
	chvRecv = chv.Convert(chtypeRecv)

	return
}

// makeSlice returns a slice of element type elemType.
func makeSlice(elemType reflect.Type) (sliceValue reflect.Value) {
	// Get the slice type
	sliceType := reflect.SliceOf(elemType)

	// Make the slice
	sliceValue = reflect.MakeSlice(sliceType, 0, 0)
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
func (ee *LocalEventEmitter) Publish(chid string, v interface{}) interface{} {
	// Get the type of v
	vtype := reflect.TypeOf(v)

	if vtype.Kind() == reflect.Ptr {
		if reflect.ValueOf(v).Elem().Type().Kind() != reflect.Struct {
			log.Panic("<Publish> Pointer must point to struct type")
		}
	} else if vtype.Kind() != reflect.Struct {
		log.Panic("<Publish> Element type must be a struct")
	}

	// Make directed channels

	// TODO Refactor if performance is an issue. The channels/slice does not need to be constructed in all cases.
	chvSend, chvRecv := makeSendRecv(vtype)
	slicev := makeSlice(chvSend.Type())

	event, ok := ee.eventMap[chid]

	if ok {
		if event.ElemType != vtype {
			log.Panic("<Publish> Tried to publish on a existing event with different element types")
		}

		// Check if the event was created by a publisher
		if event.NumPublishers > 0 {
			event.NumPublishers += 1
			return event.PublSend.Interface()
		}
	} else {
		// Create a new event and store in map
		ee.eventMap[chid] = &Event{
			ElemType:      vtype,
			PublSend:      reflect.Value{}, // Not assigned yet
			PublRecv:      reflect.Value{}, // Not assigned yet
			Subscribers:   slicev,
			NumPublishers: 0,
		}

		event = ee.eventMap[chid]
	}

	// The event exists at this point with the publisher channels missing
	ee.eventMap[chid].PublSend = chvSend
	ee.eventMap[chid].PublRecv = chvRecv
	ee.eventMap[chid].NumPublishers += 1

	return chvSend.Interface()
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
func (ee *LocalEventEmitter) Subscribe(chid string, v interface{}) interface{} {
	// Subscribe returns a read-only channel of element type of v

	// Get the type of v
	vtype := reflect.TypeOf(v)

	if vtype.Kind() == reflect.Ptr {
		if reflect.ValueOf(v).Elem().Type().Kind() != reflect.Struct {
			log.Panic("<Subscribe> Pointer must point to struct type")
		}
	} else if vtype.Kind() != reflect.Struct {
		log.Panic("<Subscribe> Element type must be a struct")
	}

	// Make directed channels
	chvSend, chvRecv := makeSendRecv(vtype)

	event, ok := ee.eventMap[chid]
	// Check if element is present
	if ok {
		if event.ElemType != vtype {
			log.Panic("<Subscribe> Tried to subscribe on a existing event with different element type")
		}

	} else {
		// There are no events with this identifier, creates one
		slicev := makeSlice(chvSend.Type())

		// Create a new event and store in map
		ee.eventMap[chid] = &Event{
			ElemType:      vtype,
			PublSend:      reflect.Value{}, // Not assigned yet
			PublRecv:      reflect.Value{}, // Not assigned yet
			Subscribers:   slicev,
			NumPublishers: 0,
		}

		event = ee.eventMap[chid]
	}

	// Append write only channel
	event.Subscribers = reflect.Append(event.Subscribers, chvSend)

	// Return read only channel
	return chvRecv.Interface()

}

// NewLocalEventEmitter is the constructor of LocalEventEmitter
func NewLocalEventEmitter() *LocalEventEmitter {
	return &LocalEventEmitter{
		eventMap: make(map[string]*Event),
		closed:   false,
	}
}

// Run method listens on publishing channels and broadcasts data to subscribers.
func (ee *LocalEventEmitter) Run() {
	ee.closed = false

	for !ee.IsClosed() {

		// TODO Use a priority queue instead of linear range over the map?
		for _, event := range ee.eventMap {

			chv := event.PublRecv

			cases := []reflect.SelectCase{
				reflect.SelectCase{
					reflect.SelectRecv,
					chv,
					reflect.ValueOf(nil),
				},
				reflect.SelectCase{
					reflect.SelectDefault,
					reflect.ValueOf(nil),
					reflect.ValueOf(nil),
				},
			}

			chosen, recv, _ := reflect.Select(cases)
			switch chosen {
			case 0:

				/* // FIXME This following code can be used for inspecting data
				for i := 0; i < recv.NumField(); i++ {
					recvueField := recv.Field(i)
					typeField := recv.Type().Field(i)

					//log.Printf("Field Name: %s,\t Field Value: %v\n", typeField.Name, recvueField.Interface())
				}
				*/
				if event.Subscribers.Type().Kind() != reflect.Slice {
					log.Panic("event.Subscribers is not a slice")
				}

				for i := 0; i < event.Subscribers.Len(); i++ {
					sub := event.Subscribers.Index(i)
					sub.Send(recv)
				} // end for
			case 1:
				// Nothing received
				break
			}
		} // end for
	} // end if
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

// Close will close all open channels to subscribers.
func (ee *LocalEventEmitter) Close() error {
	if ee.IsClosed() {
		return errors.New("Can not close already closed event emitter")
	}
	// Close the subscriber channels
	for _, event := range ee.eventMap {
		for i := 0; i < event.Subscribers.Len(); i++ {
			event.Subscribers.Index(i).Close()
		}
		//delete(ee.eventMap, k)
	}

	// Clear the event map
	for k := range ee.eventMap {
		delete(ee.eventMap, k)
	}

	ee.closed = true
	return nil
}

// IsClosed returns true if Close method has been called.
//
// Call Run method again to open
func (ee *LocalEventEmitter) IsClosed() bool {
	return ee.closed
}
