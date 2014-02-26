package aioli

import (
	//"encoding/gob"
	"encoding/json"
	"errors"
	"github.com/maxnordlund/breamio/briee"
	"log"
	"reflect"
	"time"
)

type BasicIOManager struct {
	EEMap map[int]*briee.EventEmitter
	// Add publisher channels map[eventID](chan<-interface())
}

func NewBasicIOManager() *BasicIOManager {
	return &BasicIOManager{make(map[int]*briee.EventEmitter)}
}

func (biom *BasicIOManager) Listen(recvCh <-chan ExtPkg) {
	// Listen on incomming data of ExtPkg data
	select {
	case recvData := (<-recvCh):
		//log.Printf("Recv data %v", recvData)

		// TODO Check if recvData.ID == 0, if so send on all emitters
		// TODO Check if an publisher channel already exists for <ee_id, event_name> pair

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

	default:
		log.Printf("No data, sleep...")
		time.Sleep(1 * time.Second)
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
