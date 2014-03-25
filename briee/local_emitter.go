package briee

import (
	"errors"
	//"log"
	"reflect"
	"sync"
)

type Event struct {
	ElemType     reflect.Type   // Underlying element type
	PublisherWG  sync.WaitGroup // Number of publishers active
	DataChan     reflect.Value  // Internal data channel
	Subscribers  reflect.Value  // List of write-only channels to subscribers
	CanPublish   bool
	CanSubscribe bool
	ChannelReady chan bool
	// TODO Additional map for unsubscribing
}

func newEvent(elemtype reflect.Type) *Event {
	return &Event{
		ElemType: elemtype,
		//PublisherWG:
		DataChan:     reflect.Value{},
		Subscribers:  reflect.Value{},
		CanPublish:   false,
		CanSubscribe: false,
		ChannelReady: make(chan bool, 1), // Buffered with one, important
	}
}

type LocalEventEmitter struct {
	eventMap map[string]*Event
}

func newLocalEventEmitter() *LocalEventEmitter {
	return &LocalEventEmitter{
		eventMap: make(map[string]*Event),
	}
}

func (ee *LocalEventEmitter) Publish(eventID string, v interface{}) interface{} {
	vtype := reflect.TypeOf(v)

	// Check if event is existing
	event, ok := ee.eventMap[eventID]
	if ok {
		// Event exisits, check that element types are consistent
		if event.ElemType != vtype {
			panic("Cannot publish on an existing event with different element types")
		}
	} else {
		// No existing event, create one and store in map
		event = newEvent(vtype)
		ee.eventMap[eventID] = event
	}
	// Create the write and read channels
	sendChan, recvChan := makeDirChannels(v, 0)

	// Need to add 1 to the waitgroup to get the overhead function to work
	event.PublisherWG.Add(1)

	// Check if able to add publisher
	if !event.CanPublish {
		// Create the internal data channel
		event.DataChan = makeChan(v)
		event.ChannelReady <- true
		event.CanPublish = true

		go func() {
			event.PublisherWG.Wait()
			event.DataChan.Close()
			event.CanPublish = false
		}()
	}

	go func() {
		for {
			if data, okRecv := recvChan.Recv(); okRecv {
				event.DataChan.TrySend(data)
			} else {
				break
			}
		}
		event.PublisherWG.Done()
	}()

	return sendChan.Interface()
}

func (ee *LocalEventEmitter) Subscribe(eventID string, v interface{}) interface{} {
	// Check if event is existing
	vtype := reflect.TypeOf(v)

	// Check if event is existing
	event, ok := ee.eventMap[eventID]
	if ok {
		// Event exisits, check that element types are consistent
		if event.ElemType != vtype {
			panic("Cannot subscribe on an existing event with different element types")
		}
	} else {
		// No existing event, create one and store in map
		event = newEvent(vtype)
		ee.eventMap[eventID] = event
	}

	// Create the write and read channels
	sendChan, recvChan := makeDirChannels(v, 256)

	if !event.CanSubscribe {

		event.Subscribers = makeSlice(v)
		event.CanSubscribe = true
		go func() {
			<-event.ChannelReady

			for {
				//if data, chanOpen := event.DataChan.TryRecv(); data.IsValid() {
				if data, ok := event.DataChan.Recv(); ok {
					for i := 0; i < event.Subscribers.Len(); i++ {
						ch := event.Subscribers.Index(i)
						ch.TrySend(data)
					}
				} else {
					// Clean up
					for i := 0; i < event.Subscribers.Len(); i++ {
						ch := event.Subscribers.Index(i)
						ch.Close()
					}
					break
				}
			}
		}()
	}

	event.Subscribers = reflect.Append(event.Subscribers, sendChan)
	return recvChan.Interface()
}

/*
func (ee *LocalEventEmitter) Dispatch (eventID string, v interface{}) {
	if event, ok := ee.eventMap[eventID]; ok {
		data := reflect.ValueOf(v)
		if data.Type() != event.ElemType {
			panic("Cannot dispatch value different from the registered type")
		}
		event.DataChan.TrySend(data)
	}
}
*/
func (ee *LocalEventEmitter) TypeOf(eventID string) (reflect.Type, error) {
	if event, ok := ee.eventMap[eventID]; ok {
		return event.ElemType, nil
	} else {
		return nil, errors.New("No event with that identifier is registred")
	}
}

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

func makeChan(v interface{}) reflect.Value {
	vtype := reflect.TypeOf(v)
	chtype := reflect.ChanOf(reflect.BothDir, vtype)
	chv := reflect.MakeChan(chtype, 256)
	return chv
}

func makeSlice(v interface{}) reflect.Value {
	vtype := reflect.TypeOf(v)
	chtype := reflect.ChanOf(reflect.SendDir, vtype) // Note SendDir
	slicetype := reflect.SliceOf(chtype)

	slicev := reflect.MakeSlice(slicetype, 0, 0)

	return slicev
}
