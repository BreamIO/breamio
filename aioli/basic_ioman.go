package aioli

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"reflect"
	"time"
)

// BasicIOManager implements IOManager.
type BasicIOManager struct {
	lookuper EmitterLookuper
	dataChan chan ExtPkg
	publMap  map[publMapEntry]*reflect.Value
	closed   bool
}

// newBasicIOManager creates a new BasicIOManager.
func newBasicIOManager(lookuper EmitterLookuper) *BasicIOManager {
	return &BasicIOManager{
		lookuper: lookuper,
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
func (biom *BasicIOManager) Listen(dec Decoder, logger *log.Logger) {
	for !biom.IsClosed() {
		var ep ExtPkg
		err := dec.Decode(&ep)
		if err != nil {
			if err == io.EOF {
				logger.Printf("Connection closed.")
				return
			} else {
				logger.Printf("Decoding error, %v", err)
				return
			}
			time.Sleep(time.Millisecond * 500)
		} else {
			logger.Println("Recieved:", ep)
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
	if ee, err := biom.lookuper.EmitterLookup(recvData.ID); err == nil {

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
			//log.Println(reflect.Indirect(ptr).Interface())
			if err != nil {
				log.Println(err)
			}

			ee.Dispatch(recvData.Event, ptr.Elem().Interface())

		}
	} else {
		log.Printf("No match for packet: %v", recvData)
		time.Sleep(time.Millisecond * 500)
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
