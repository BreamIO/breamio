package briee

import (
	"errors"
	//"log"
	"reflect"
	//"sync"
)

type Event struct {
	ElemType     reflect.Type   // Underlying element type
	//PublisherWG  sync.WaitGroup // Number of publishers active
	DataChan     reflect.Value  // Internal data channel
	Subscribers  reflect.Value  // List of write-only channels to subscribers
	//CanPublish   bool
	CanSubscribe bool
	//ChannelReady  chan struct{}
	SubscriberMap map[reflect.Value]reflect.Value
}

func newEvent(elemtype reflect.Type) *Event {
	return &Event{
		ElemType: elemtype,
		//DataChan:      reflect.Value{},
		DataChan:     makeChan(elemtype),
		Subscribers:  reflect.Value{},
		//CanPublish:   false,
		CanSubscribe: false,
		//ChannelReady:  make(chan struct{}), // FIXME, might need to be buffered
		SubscriberMap: make(map[reflect.Value]reflect.Value),
	}
}

type LocalEventEmitter struct {
	eventMap map[string]*Event
	open     bool
	done     chan struct{}
}

func newLocalEventEmitter() *LocalEventEmitter {
	return &LocalEventEmitter{
		eventMap: make(map[string]*Event),
		open:     true,
		done:     make(chan struct{}),
	}
}

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

func (ee *LocalEventEmitter) Subscribe(eventID string, v interface{}) interface{} {

	event := ee.event(eventID, v)

	// Create the write and read channels
	sendChan, recvChan := makeDirChannels(v, 256)

	event.SubscriberMap[recvChan] = sendChan

	if !event.CanSubscribe {

		event.Subscribers = makeSlice(v)
		event.CanSubscribe = true
		go func() {
			//<-event.ChannelReady
			for ee.IsOpen() {
				if data, ok := event.DataChan.Recv(); ok {
					for i := 0; i < event.Subscribers.Len(); i++ {
						ch := event.Subscribers.Index(i)
						ch.TrySend(data)
					}
				} else {
					break
				}
			}

			// Clean up
			for i := 0; i < event.Subscribers.Len(); i++ {
				ch := event.Subscribers.Index(i)
				ch.Close()
			}
		}()
	}

	event.Subscribers = reflect.Append(event.Subscribers, sendChan)
	return recvChan.Interface()
}

func (ee *LocalEventEmitter) Dispatch(eventID string, v interface{}) {
	event := ee.event(eventID, v)
	//event.PublisherWG.Add(1)
	//event.RunPublisherOverhead(v)
	event.DataChan.TrySend(reflect.ValueOf(v))
	// TODO Skriv direkt på DataChan!!
	//event.PublisherWG.Done()
}

func (ee *LocalEventEmitter) TypeOf(eventID string) (reflect.Type, error) {
	if event, ok := ee.eventMap[eventID]; ok {
		return event.ElemType, nil
	} else {
		return nil, errors.New("No event with that identifier is registred")
	}
}

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

func (ee *LocalEventEmitter) Wait() {
	// Wait for close to finish
	<-ee.done
}

func (ee *LocalEventEmitter) IsOpen() bool {
	return ee.open
}

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

func makeChan(vtype reflect.Type) reflect.Value {
	//vtype := reflect.TypeOf(v)
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
