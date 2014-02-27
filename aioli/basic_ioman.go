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

type ChanEntry struct {
	Event string
	ID int
}

type BasicIOManager struct {
	EEMap map[int]*briee.EventEmitter
	//Publisher map[ChanEntry]*reflect.Value // TODO
	dataChan chan ExtPkg
}

func NewBasicIOManager() *BasicIOManager {
	return &BasicIOManager{
		EEMap:  make(map[int]*briee.EventEmitter),
		dataChan:  make(chan ExtPkg),
		//EEChans map[ChanEntry]*reflect.Value // TODO
		}
}

func (biom *BasicIOManager) Listen (r io.Reader){
	// TODO make private and implement Add/Remove listeners funcionallity
	var data []byte
	var ep ExtPkg
	dec := gob.NewDecoder(r)

	for {
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
	for {
		select {
			case recvData := (<-biom.dataChan):
				biom.handle(recvData)
		}
	}
}

func (biom *BasicIOManager) handle(recvData ExtPkg) {

	if ee, ok := biom.EEMap[recvData.ID]; ok {
		// Check type with emitter
		rtype, err := (*ee).TypeOf(recvData.Event) // Note ee ptr

		if err != nil {
			log.Println(err)

		} else {
			// Decode data as json according to rtype reflect.Type from event emitter
			// Use a provided decoder, but at this moment json is a hardcoded selection

			zeroValInterface := reflect.Zero(rtype).Interface()

			// publCh is a write only channel of element type of rtype
			publCh := reflect.ValueOf((*ee).Publish(recvData.Event, zeroValInterface))

			buf := recvData.Data      // json format
			ptr := reflect.New(rtype) // New value of that type

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
