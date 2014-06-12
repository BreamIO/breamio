package aioli

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"os"
	"reflect"
	"time"
)

// BasicIOManager implements IOManager.
type BasicIOManager struct {
	lookuper EmitterLookuper
	dataChan chan ExtPkg
	publMap  map[publMapEntry]*reflect.Value
	closed   bool
	logger   *log.Logger
}

// newBasicIOManager creates a new BasicIOManager.
func newBasicIOManager(lookuper EmitterLookuper) *BasicIOManager {
	return &BasicIOManager{
		lookuper: lookuper,
		dataChan: make(chan ExtPkg),
		publMap:  make(map[publMapEntry]*reflect.Value),
		closed:   true,
		logger:   log.New(os.Stderr, "[Aioli]", log.LstdFlags),
	}
}

type publMapEntry struct {
	Event string
	ID    int
}

// Listen will try to decode ExtPkg structs from the underlying data stream of the provided decoder and handle the structs accordingly.
//
// Requires that the IOManager Run method is running.
func (biom *BasicIOManager) Listen(codec EncodeDecoder, logger *log.Logger) {
	biom.logger = logger
	for !biom.IsClosed() {
		var ep ExtPkg
		err := codec.Decode(&ep)
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
			if ep.Subscribe {
				//logger.Println("Recieved subscription request for", ep.Event)
				go biom.handleSubscription(ep, codec, logger)
				continue
			}
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
			biom.logger.Println("TypeOf error:", err)
		} else {
			// Decode data as json according to rtype reflect.Type from event emitter
			buf := recvData.Data      // buf is of encoded json format
			ptr := reflect.New(rtype) // New value of that wanted type

			// TODO Replace json.Unmarshal with provided decoder
			err := json.Unmarshal(buf, ptr.Interface()) // Unmarshal into the interface of the pointer
			//biom.logger.Println(reflect.Indirect(ptr).Interface())
			if err != nil {
				biom.logger.Println("JSON error:", err)
			}

			ee.Dispatch(recvData.Event, ptr.Elem().Interface())
		}
	} else {
		time.Sleep(time.Millisecond * 500)
		biom.logger.Printf("No match for packet: %v", recvData)
	}
}

func (biom *BasicIOManager) handleSubscription(recvData ExtPkg, enc Encoder, logger *log.Logger) {
	ee, err := biom.lookuper.EmitterLookup(recvData.ID)
	if err != nil {
		logger.Printf("Subscription for event \"%s\" failed: No such emitter %d.\n", recvData.Event, recvData.ID)
		enc.Encode(ExtPkg{
			Event:     recvData.Event,
			Subscribe: true,
			ID:        recvData.ID,
			Error:     NewError("No such emitter"),
		})
		return
	}

	rtype, err := ee.TypeOf(recvData.Event) // Note ee ptr
	if err != nil {
		logger.Printf("Subscription for event \"%s\" failed: No such event.\n", recvData.Event)
		enc.Encode(ExtPkg{
			Event:     recvData.Event,
			Subscribe: true,
			ID:        recvData.ID,
			Error:     NewError("No such event"),
		})
		return
	}

	template := reflect.New(rtype).Elem().Interface()
	dataCh := reflect.ValueOf(ee.Subscribe(recvData.Event, template)) //Reflected channel.
	// No we do not care about exact type.

	logger.Printf("Subscription for event \"%s\" on emitter %d started.\n", recvData.Event, recvData.ID)
	for !biom.IsClosed() {
		val, ok := dataCh.TryRecv()
		if !ok {
			continue
		}

		if val.IsValid() {
			//Now we are clear to do stuff with data.
			data, err := json.Marshal(val.Interface())
			if err != nil {
				logger.Printf("Subscription for event \"%s\" encountered a error during encoding of payload: %s.\n", recvData.Event, err.Error())
				logger.Println(val.Interface())
				return
			}
			err = enc.Encode(ExtPkg{
				Event:     recvData.Event,
				Subscribe: true,
				ID:        recvData.ID,
				Data:      data,
			})
			if err != nil {
				logger.Printf("Subscription for event \"%s\" encountered a error during encoding: %s.\n", recvData.Event, err.Error())
				return
			}
		}
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

func (biom *BasicIOManager) Dispatch(ep ExtPkg) {
	biom.dataChan <- ep
}
