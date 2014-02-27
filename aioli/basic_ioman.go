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

type BasicIOManager struct {
	EEMap map[int]*briee.EventEmitter
	dataChan chan ExtPkg
	//publChans map[string]*reflect.Value // TODO
}

func NewBasicIOManager() *BasicIOManager {
	return &BasicIOManager{
		EEMap:  make(map[int]*briee.EventEmitter),
		dataChan:  make(chan ExtPkg),
		// publChans: make(map[string]*reflect.Value), // TODO Event + string(ID) as key
		}
}

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

func (biom *BasicIOManager) Run() {
	// Run listens on the internal channel of ExtPkg data on which all listerners send data on
	for {
		select {
			case recvData := (<-biom.dataChan):
				biom.handle(recvData)
		}
	}
}

func (biom *BasicIOManager) handle(recvData ExtPkg) {
	// Handle tries to decode and send the provided ExtPkg on one or more event emitters

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

func (biom *BasicIOManager) AddEE(ee *briee.EventEmitter, id int) error {
	// Add an pointer to an event emitter in not already present, then return error
	if _, ok := biom.EEMap[id]; !ok {
		biom.EEMap[id] = ee
		return nil
	} else {
		return errors.New("Can not add event emitter with existing identifier")
	}
}

func (biom *BasicIOManager) RemoveEE(id int) error {
	// Remove the pointer to an event emitter with id if present, if not return error
	if _, ok := biom.EEMap[id]; ok {
		delete(biom.EEMap, id)
		return nil
	} else {
		return errors.New("Can not remove non-existing event emitter")
	}
}
