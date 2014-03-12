package aioli

import (
	"encoding/json"
	"errors"
	"github.com/maxnordlund/breamio/briee"
	"log"
	"reflect"
	"time"
	"io"
)

// BasicIOManager implements IOManager.
type BasicIOManager struct {
	eeMap    map[int]briee.EventEmitter
	dataChan chan ExtPkg
	publMap  map[publMapEntry]*reflect.Value
	closed   bool
}

// NewBasicIOManager creates a new BasicIOManager.
func newBasicIOManager() *BasicIOManager {
	return &BasicIOManager{
		eeMap:    make(map[int]briee.EventEmitter),
		dataChan: make(chan ExtPkg),
		publMap:  make(map[publMapEntry]*reflect.Value),
		closed:   true,
	}
}

type publMapEntry struct {
	Event string
	ID    int
}

// Listen will try to decode ExtPkg structs from the underlying data stream of the provided decoder and handle the structs accordingly.
//
// Requires that the IOManager Run method is running.
func (biom *BasicIOManager) Listen(dec Decoder) {
	for !biom.IsClosed() {
		var ep ExtPkg
		err := dec.Decode(&ep)
		if err != nil {
			if err == io.EOF {
				log.Printf("EOF!")
				return
			} else {
				log.Printf("Decoding error, %v", err)
				return
			}
			time.Sleep(time.Millisecond * 500)
		} else {
			biom.dataChan <- ep
		}
	}
}

// Run listens on the internal channel of ExtPkg data on which all listerners send data on.
func (biom *BasicIOManager) Run() {
	biom.closed = false
	for !biom.IsClosed() {
		select {
		case recvData := (<-biom.dataChan):
			biom.handle(recvData)
		}
	}
}

// Handle tries to decode and send the provided ExtPkg on one or more event emitters
func (biom *BasicIOManager) handle(recvData ExtPkg) {
	// TODO Add broadcast functionality
	if ee, ok := biom.eeMap[recvData.ID]; ok {

		// Look up the type in the event emitter
		rtype, err := ee.TypeOf(recvData.Event) // Note ee ptr

		if err != nil {
			log.Println(err)

		} else {

			// Decode data as json according to rtype reflect.Type from event emitter
			buf := recvData.Data      // buf is of encoded json format
			ptr := reflect.New(rtype) // New value of that wanted type

			// TODO Replace json.Unmarshal with provided decoder
			err := json.Unmarshal(buf, ptr.Interface()) // Unmarshal into the interface of the pointer
			if err != nil {
				log.Println(err)
			}

			if pChanPtr, ok := biom.publMap[publMapEntry{Event: recvData.Event, ID: recvData.ID}]; ok {
				(*pChanPtr).Send(ptr.Elem())

			} else {
				zeroValInterface := reflect.Zero(rtype).Interface()
				// publCh is a write only channel of element type of rtype
				publCh := reflect.ValueOf(ee.Publish(recvData.Event, zeroValInterface))

				// Save the publisher channel for future use
				biom.publMap[publMapEntry{Event: recvData.Event, ID: recvData.ID}] = &publCh
				publCh.Send(ptr.Elem()) // Send decoded element on channel
			}

		}
	} else {
		log.Printf("No match for packet: %v", recvData)
		time.Sleep(time.Millisecond * 500)
	}
}

// AddEE adds a pointer to an event emitter and an identifier if not already present. Will return a error if unsuccessful.
//
// Provided interger identigier must not be zero as this is reverved for broadcasting all event emitters. Doing so will return an error.
func (biom *BasicIOManager) AddEE(ee briee.EventEmitter, id int) error {
	if id == 0 {
		return errors.New("Integer identifier zero is reserved for broadcasting")
	}
	if _, ok := biom.eeMap[id]; !ok {
		biom.eeMap[id] = ee
		return nil
	} else {
		return errors.New("Can not add event emitter with existing identifier")
	}
}

// RemoveEE removes the registered event emitter if the provided identifier is present. Will return a error if unsuccessful.
//
// Provided interger identigier must not be zero as this is reverved for broadcasting all event emitters. Doing so will return an error.
func (biom *BasicIOManager) RemoveEE(id int) error {
	if id == 0 {
		return errors.New("Integer identifier zero is reserved for broadcasting")
	}
	if ee, ok := biom.eeMap[id]; ok {
		err := ee.Close()
		if err != nil {
			return err
		}

		delete(biom.eeMap, id)
		return nil
	} else {
		return errors.New("Can not remove non-existing event emitter")
	}
}

// Close will cause the Run method and all running listeners to terminate.
//
// Will return an error if the IOManager is not running and cannot be closed.
// Will also return an error if already closed.
func (biom *BasicIOManager) Close() error {
	if biom.IsClosed() {
		return errors.New("Can not close already closed IOManager")
	}
	biom.closed = true
	return nil
}

// IsClosed returns true if IOManager is currently closed.
//
// Call Run method to open.
func (biom *BasicIOManager) IsClosed() bool {
	return biom.closed
}
