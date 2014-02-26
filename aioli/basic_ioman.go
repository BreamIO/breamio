package aioli

import (
	//"encoding/gob"
	//"encoding/json"
	"log"
	"time"
	"errors"
	//"reflect"
	"github.com/maxnordlund/breamio/briee"
)

type BasicIOManager struct {
	EEMap map[int]*briee.EventEmitter
}

func NewBasicIOManager() *BasicIOManager {
	return &BasicIOManager{make(map[int]*briee.EventEmitter)}
}

func (biom * BasicIOManager) Listen (recvCh <-chan ExtPkg) {
	// Listen on incomming data
	select{
		case recvData := (<-recvCh):
			log.Printf("Recv data %v", recvData)
			// Check type with emitter
			/*
			rtype, err := biom.EEMap[recvData.ID].TypeOf(recvData)
			if err != nil {
				log.Printf(err)
			} else {
				// Parse recvData.Data as rtype reflect.Value, using gob.Decoder?
				// If successful, send value.interface() on emitter
			}
			*/
		default:
			log.Printf("No data, sleep...")
			time.Sleep(1 * time.Second)
	}

}

func (biom * BasicIOManager) AddEE (ee *briee.EventEmitter, id int) error {
	// Add an pointer to an event emitter in not already present, then return error
	if _, ok := biom.EEMap[id]; !ok {
		biom.EEMap[id] = ee
		return nil
	} else {
		return errors.New("Can not add event emitter with existing identifier")
	}
}

func (biom * BasicIOManager) RemoveEE (id int) error {
	// Remove the pointer to an event emitter with id if present, if not return error
	if _, ok := biom.EEMap[id]; ok {
		delete(biom.EEMap, id)
		return nil
	} else {
		return errors.New("Can not remove non-existing event emitter")
	}
}
