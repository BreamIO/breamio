// Briee defines and implements the interface Emitter
package briee

import (
	"log"
	"reflect"
)

type EventEmitter interface {
	Publish(chid string, v interface{}) interface{}
	Subscribe(chid string, v interface{}) interface{}
	Run() // Creates and runs the emitter
}

type Event struct {
	ElemType      reflect.Type // Element type
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
	//log.Printf("<Publish> Type of chtype: %v\n", chtype)

	// Create the directed channel types
	chtypeSend := reflect.ChanOf(reflect.SendDir, vtype)
	chtypeRecv := reflect.ChanOf(reflect.RecvDir, vtype)

	if !chtype.ConvertibleTo(chtypeSend) {
		//log.Printf("<Publish> I can not convert to one way channel... \n")
	}
	if !chtype.ConvertibleTo(chtypeRecv) {
		//log.Printf("<Publish> I can not convert to one way channel... \n")
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
	// Publish creates a new Event with the string identifier chid.
	// If Publish is called with an already existing Event, Publish will
	// return the already created Publisher channel unless the types are
	// inconsistent and will cause a panic.

	// Get the type of v
	vtype := reflect.TypeOf(v)
	//log.Printf("<Publish> Type of vtype: %v\n", vtype)

	if vtype.Kind() != reflect.Struct {
		log.Panic("<Publish> Element type must be a struct")
	}

	// Make directed channels
	chvSend, chvRecv := makeSendRecv(vtype) // TODO Refactor to after the if statements
	slicev := makeSlice(chvSend.Type())     // TODO Refactor to after the if statements

	event, ok := e.eventMap[chid]

	if ok {
		// This kind of event already exists
		// But we don't know if they have the same elem type
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

	// The event exists at this point with the publchans missing
	e.eventMap[chid].PublSend = chvSend
	e.eventMap[chid].PublRecv = chvRecv
	e.eventMap[chid].NumPublishers += 1

	return chvSend.Interface()
}

func (e *LocalEventEmitter) Subscribe(chid string, v interface{}) interface{} {
	// Create a channel of type v
	// Return the read-only channel and save the write-only in map

	// Get the type of v
	vtype := reflect.TypeOf(v)
	//log.Printf("<Subscribe> Type of vtype: %v\n", vtype)

	if vtype.Kind() != reflect.Struct {
		log.Panic("<Publish> Element type must be a struct")
	}

	// Make directed channels
	chvSend, chvRecv := makeSendRecv(vtype)

	event, ok := e.eventMap[chid]
	// Check if element is present
	if ok {
		// This kind of event already exists
		// but we don't know if they have the same elem type
		if event.ElemType != vtype {
			log.Panic("<Subscribe> Tried to subscribe on a existing event with different element type")
		}

	} else {
		// There are no events present, create one
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

	// append write only channel
	event.Subscribers = reflect.Append(event.Subscribers, chvSend)

	// return read only channel
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

			//chosen, recv, recvOK := reflect.Select(cases)
			chosen, recv, _ := reflect.Select(cases)
			switch chosen {
			case 0:
				////log.Printf("Got recv: %v\n", recv)

				/*
					for i := 0; i < recv.NumField(); i++ {
						recvueField := recv.Field(i)
						typeField := recv.Type().Field(i)

						//log.Printf("Field Name: %s,\t Field Value: %v\n", typeField.Name, recvueField.Interface())
					}
				*/
				if event.Subscribers.Type().Kind() != reflect.Slice {
					log.Panic("event.Subscribers is not a struct!")
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
