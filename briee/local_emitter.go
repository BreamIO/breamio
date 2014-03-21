package briee

import (
	"errors"
	"log"
	"reflect"
	"sync"
)

type Event struct {
	ElemType     reflect.Type   // Underlying element type
	PublisherWG  sync.WaitGroup // Number of publishers active
	DataChan     interface{}    // Internal data channel
	Subscribers  interface{}    // List of write-only channels to subscribers
	CanPublish   bool
	CanSubscribe bool
	ChannelReady chan bool
	// TODO Additional map for unsubscribing
}

func newEvent(elemtype reflect.Type) *Event {
	return &Event{
		ElemType: elemtype,
		//PublisherWG:
		DataChan:     nil,
		Subscribers:  nil,
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
			panic("Cannot not publish on an existing event with different element types")
		}
	} else {
		// No existing event, create one
		event = newEvent(vtype)
	}
	// Create the write and read channels
	sendChan, recvChan := makeDirChannels(v)

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
			close(event.DataChan.(chan interface{}))
			event.CanPublish = false
		}()
	}

	go func() {
		for data := range recvChan.(chan interface{}) {
			event.DataChan.(chan interface{}) <- data
		}
	}()

	return sendChan
}

func (ee *LocalEventEmitter) Subscribe(eventID string, v interface{}) interface{} {
	// TODO REFACTOR BEGIN, move to other method, because of code-copy
	// Check if event is existing
	vtype := reflect.TypeOf(v)

	// Check if event is existing
	event, ok := ee.eventMap[eventID]
	if ok {
		// Event exisits, check that element types are consistent
		if event.ElemType != vtype {
			panic("Cannot not subscribe on an existing event with different element types")
		}
	} else {
		// No existing event, create one
		event = newEvent(vtype)
	}
	// TODO REFACTOR END

	// Create the write and read channels
	sendChan, recvChan := makeDirChannels(v)

	if !event.CanSubscribe {
		// FIXME slice initilization
		event.Subscribers = makeSlice(v)
		go func() {
			<-event.ChannelReady
			for data := range event.DataChan.(chan interface{}) {
				for ch := range event.Subscribers.([]chan<- interface{}) {
					log.Printf("Data: %v\tCh: %v", reflect.ValueOf(data), reflect.ValueOf(ch))
					/*
						select {
						case ch <- data:
							// Done
							log.Printf("Successfully sent data")
						default:
							log.Printf("Unable to send data")
						}
					*/
				}
			}
		}()
	}

	//event.Subscribers = append(event.Subscribers.([]chan<- interface{}), sendChan)
	reflect.Append(reflect.ValueOf(event.Subscribers), reflect.ValueOf(sendChan))

	return recvChan
}

func (ee *LocalEventEmitter) TypeOf(eventID string) (reflect.Type, error) {
	if event, ok := ee.eventMap[eventID]; ok {
		return event.ElemType, nil
	} else {
		return nil, errors.New("No event with that identifier is registred")
	}
}

func makeDirChannels(v interface{}) (interface{}, interface{}) {

	vtype := reflect.TypeOf(v)

	// Get channel type
	chtype := reflect.ChanOf(reflect.BothDir, vtype)

	// Create the directed channel types
	chtypeSend := reflect.ChanOf(reflect.SendDir, vtype)
	chtypeRecv := reflect.ChanOf(reflect.RecvDir, vtype)

	// Make a two-way channel
	chv := reflect.MakeChan(chtype, 0)

	// Make a write-only channel
	chvSend := chv.Convert(chtypeSend)

	// Make a read-only channel
	chvRecv := chv.Convert(chtypeRecv)

	return chvSend.Interface(), chvRecv.Interface()

}

func makeChan(v interface{}) interface{} {
	vtype := reflect.TypeOf(v)
	chtype := reflect.ChanOf(reflect.BothDir, vtype)
	chv := reflect.MakeChan(chtype, 0)
	return chv.Interface()
	//return (chan interface{})(chv)
	//return chv.Interface().(chan interface{})
	//return chv.Interface().(chan reflect.TypeOf(v))
	//return chv.Interface()
	//return chv
}

func makeSlice(v interface{}) interface{} {
	vtype := reflect.TypeOf(v)
	chtype := reflect.ChanOf(reflect.SendDir, vtype) // Note SendDir
	slicetype := reflect.SliceOf(chtype)

	slicev := reflect.MakeSlice(slicetype, 0, 0)

	return slicev.Interface()
}
