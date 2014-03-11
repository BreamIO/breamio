package aioli

import (
	//"encoding/gob"
	"encoding/json"
	"errors"
	"github.com/maxnordlund/breamio/briee"
	//"io"
	"log"
	"reflect"
	//"time"
)

// BasicIOManager implements IOManager.
type BasicIOManager struct {
	eeMap    map[int]briee.EventEmitter
	dataChan chan ExtPkg
	publMap  map[publMapEntry]*reflect.Value
}

// NewBasicIOManager creates a new BasicIOManager.
func newBasicIOManager() *BasicIOManager {
	return &BasicIOManager{
		eeMap:    make(map[int]briee.EventEmitter),
		dataChan: make(chan ExtPkg),
		publMap:  make(map[publMapEntry]*reflect.Value),
	}
}

type publMapEntry struct {
	Event string
	ID    int
}

//func (biom *BasicIOManager) Listen(r io.Reader) {
//	// TODO make private and implement Add/Remove listeners funcionallity
//	var data []byte
//	var ep ExtPkg
//	dec := gob.NewDecoder(r)
//
//	for { // inf loop
//		err := dec.Decode(&data)
//		if err != nil {
//			log.Printf("Decoding failed, sleep ...")
//			time.Sleep(50 * time.Millisecond)
//			continue
//		}
//
//		err = json.Unmarshal(data, &ep)
//		if err != nil {
//			log.Printf("Unmarshal error, ", err)
//		}
//
//		biom.dataChan <- ep
//	}
//}

// Listen will listen for ExtPkg data on the provided io.Reader and redirect for further handling.
func (biom *BasicIOManager) Listen(dec Decoder) {
	// TODO make private and implement Add/Remove listeners funcionallity
	var ep ExtPkg
	for { // inf loop, FIXME
		err := dec.Decode(&ep)
		if err != nil {
			//log.Printf("Decoding failure")
		}
		biom.dataChan <- ep
	}
}

// Run listens on the internal channel of ExtPkg data on which all listerners send data on.
func (biom *BasicIOManager) Run() {
	for {
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
