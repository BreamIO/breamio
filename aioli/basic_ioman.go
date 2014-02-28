package aioli

import (
	"encoding/gob"
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
	EEMap map[int]*briee.EventEmitter
	dataChan chan ExtPkg
	//publChans map[string]*reflect.Value // TODO
}

// NewBasicIOManager creates a new BasicIOManager.
func NewBasicIOManager() *BasicIOManager {
	return &BasicIOManager{
		EEMap:  make(map[int]*briee.EventEmitter),
		dataChan:  make(chan ExtPkg),
		// publChans: make(map[string]*reflect.Value), // TODO Event + string(ID) as key
		}
}

// Listen will listen for ExtPkg data on the provided io.Reader and redirect for further handling.
func (biom *BasicIOManager) Listen (r io.Reader){
	// TODO make private and implement Add/Remove listeners funcionallity
	var data []byte
	var ep ExtPkg
	dec := gob.NewDecoder(r)

	for { // inf loop
		err := dec.Decode(&data)
		if err != nil {
			log.Printf("Decoding failed, sleep ...")
			time.Sleep(50 * time.Millisecond)
			continue
		}

		err = json.Unmarshal(data, &ep)
		if err != nil {
			log.Printf("Unmarshal error, ", err)
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
	if ee, ok := biom.EEMap[recvData.ID]; ok {

		// Look up the type in the event emitter
		rtype, err := (*ee).TypeOf(recvData.Event) // Note ee ptr

		if err != nil {
			log.Println(err)

		} else {
			// Decode data as json according to rtype reflect.Type from event emitter
			// Use a provided decoder, but at this moment json is a hardcoded selection

			zeroValInterface := reflect.Zero(rtype).Interface()

			// publCh is a write only channel of element type of rtype
			publCh := reflect.ValueOf((*ee).Publish(recvData.Event, zeroValInterface))

			buf := recvData.Data      // buf is of encoded json format
			ptr := reflect.New(rtype) // New value of that wanted type

			// TODO Replace json.Unmarshal with provided decoder
			err := json.Unmarshal(buf, ptr.Interface()) // Unmarshal into the interface of the pointer
			if err != nil {
				log.Println(err)
			}

			// TODO Save this publisher channel in a map for future use
			publCh.Send(ptr.Elem()) // Send decoded element on channel

		}
	} else {
		log.Printf("No matching event: %v from event emitter\n", recvData.Event)
	}
}

// AddEE adds a pointer to an event emitter and an identifier if not already present. Will return a error if unsuccessful.
func (biom *BasicIOManager) AddEE(ee *briee.EventEmitter, id int) error {
	if _, ok := biom.EEMap[id]; !ok {
		biom.EEMap[id] = ee
		return nil
	} else {
		return errors.New("Can not add event emitter with existing identifier")
	}
}

// RemoveEE removes the registered event emitter if the provided identifier is present. Will return a error if unsuccessful.
func (biom *BasicIOManager) RemoveEE(id int) error {
	if _, ok := biom.EEMap[id]; ok {
		delete(biom.EEMap, id)
		return nil
	} else {
		return errors.New("Can not remove non-existing event emitter")
	}
}
