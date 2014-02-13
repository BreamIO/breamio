// Briee defines and implements the interface Emitter
package briee

import (
	"log"
	"reflect"
)

type EventEmitter interface {
	Publish(chid string, v interface{}) interface{}
	Subscribe(chid string, v interface{}) interface{}
	Run() // Runs the emitter
}

type Event struct {
	ElemType      reflect.Type	// Element type
	PublSend      reflect.Value // Write only channel
	PublRecv      reflect.Value // Read only channel
	Subscribers   reflect.Value // Slice of wrtie only channels
	NumPublishers int           // Number of publishers
}

// LocalEventEmitter implements EventEmitter
type LocalEventEmitter struct {
	eventMap map[string]*Event
}

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

func makeSlice(elemType reflect.Type) (sliceValue reflect.Value) {
	// Get the slice type
	sliceType := reflect.SliceOf(elemType)

	// Make the slice
	sliceValue = reflect.MakeSlice(sliceType, 0, 0)
	return
}

func (e *LocalEventEmitter) Publish(chid string, v interface{}) interface{} {
	// Publish returns a write-only channel with element type of parameter v

	// Get the type of v
	vtype := reflect.TypeOf(v)

	if vtype.Kind() != reflect.Struct {
		log.Panic("<Publish> Element type must be a struct")
	}

	// Make directed channels

	// TODO Refactor if performance is an issue. The channels/slice does not need to be constructed in all cases.
	chvSend, chvRecv := makeSendRecv(vtype)
	slicev := makeSlice(chvSend.Type())

	event, ok := e.eventMap[chid]

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
		e.eventMap[chid] = &Event{
			ElemType:      vtype,
			PublSend:      reflect.Value{}, // Not assigned yet
			PublRecv:      reflect.Value{}, // Not assigned yet
			Subscribers:   slicev,
			NumPublishers: 0,
		}

		event = e.eventMap[chid]
	}

	// The event exists at this point with the publisher channels missing
	e.eventMap[chid].PublSend = chvSend
	e.eventMap[chid].PublRecv = chvRecv
	e.eventMap[chid].NumPublishers += 1

	return chvSend.Interface()
}

func (e *LocalEventEmitter) Subscribe(chid string, v interface{}) interface{} {
	// Subscribe returns a read-only channel of element type of v

	// Get the type of v
	vtype := reflect.TypeOf(v)

	if vtype.Kind() != reflect.Struct {
		log.Panic("<Publish> Element type must be a struct")
	}

	// Make directed channels
	chvSend, chvRecv := makeSendRecv(vtype)

	event, ok := e.eventMap[chid]
	// Check if element is present
	if ok {
		if event.ElemType != vtype {
			log.Panic("<Subscribe> Tried to subscribe on a existing event with different element type")
		}

	} else {
		// There are no events with this identifier, creates one
		slicev := makeSlice(chvSend.Type())

		// Create a new event and store in map
		e.eventMap[chid] = &Event{
			ElemType:      vtype,
			PublSend:      reflect.Value{}, // Not assigned yet
			PublRecv:      reflect.Value{}, // Not assigned yet
			Subscribers:   slicev,
			NumPublishers: 0,
		}

		event = e.eventMap[chid]
	}

	// Append write only channel
	event.Subscribers = reflect.Append(event.Subscribers, chvSend)

	// Return read only channel
	return chvRecv.Interface()

}

func NewEventEmitter() *LocalEventEmitter {
	return &LocalEventEmitter{make(map[string]*Event)}
}

func (e *LocalEventEmitter) Run() {
	for { // infinite loop
		for _, event := range e.eventMap {

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
