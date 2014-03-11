package aioli

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"log"
	"sync"
	//"time"
	"github.com/maxnordlund/breamio/briee"
	"testing"
	//"reflect"
)

type Payload struct {
	X, Y float64
}

func TestAddRemoveEmitters(t *testing.T) {
	ee := briee.NewEventEmitter()
	ioman := NewIOManager()

	// Add event emitter
	err := ioman.AddEE(ee, 1)
	if err != nil {
		t.Errorf("Unable to add event emitter")
	}

	// Remove just added event emitter
	err = ioman.RemoveEE(1)
	if err != nil {
		t.Errorf("Unable to remove event emitter")
	}
}

func TestAddEEBC(t *testing.T) {
	ee := briee.NewEventEmitter()
	ioman := NewIOManager()

	// Add event emitter
	err := ioman.AddEE(ee, 0)
	if err == nil {
		t.Errorf("Should not be able to add broadcast identifier")
	}
}

func TestRemEEBC(t *testing.T) {
	ee := briee.NewEventEmitter()
	ioman := NewIOManager()

	// Add event emitter
	err := ioman.AddEE(ee, 1)
	if err != nil {
		t.Errorf("Unable to add event emitter")
	}

	// Remove just added event emitter
	err = ioman.RemoveEE(0)
	if err == nil {
		t.Errorf("Should not be able to remove broadcast identifier")
	}
}

func TestIOman(t *testing.T) {
	// Set up emitter
	ee := briee.NewEventEmitter()
	go ee.Run()
	subscriber := ee.Subscribe("event data", Payload{}).(<-chan Payload)

	// Set up IO manager
	ioman := NewIOManager()

	// Add event emitter
	err := ioman.AddEE(ee, 1)
	if err != nil {
		t.Errorf("Unable to add event emitter")
	}

	var network bytes.Buffer // Stand-in for the network

	// Example data from an external source
	plSend := Payload{
		X: 0.1,
		Y: 0.2,
	}

	var plRecv Payload

	var wg sync.WaitGroup
	wg.Add(2)

	go ioman.Run()
	dec := NewDecoder(&network)
	// Listen fix, TODO Clean up
	//go ioman.Listen(&network)
	go ioman.Listen(dec)

	go func() {
		// Send decodes and sends payload data on network
		send(plSend, &network)
		send(plSend, &network)
		wg.Done()
	}()

	go func() {
		// Listen on subscriber channel
		plRecv = <-subscriber
		plRecv = <-subscriber
		wg.Done()
	}()

	wg.Wait()

	if plSend != plRecv {
		t.Errorf("Got %v, want %v\n", plRecv, plSend)
	}
}

func send(pl Payload, network *bytes.Buffer) {
	// Encode Data, representing the other side, e.g. WEB, CLI

	// Encode the payload data of type Payload as json
	jsonpl, plerr := json.Marshal(pl)

	if plerr != nil {
		log.Panic("Marshal error, ", plerr)
	}

	// Construct the external package with the encoded payload
	rawPkg := ExtPkg{
		Event: "event data",
		ID:    1,
		Data:  jsonpl,
	}

	// Encode the external package as json
	jsonPkg, pkgerr := json.Marshal(rawPkg)

	if pkgerr != nil {
		log.Panic("Marshal error, ", pkgerr)
	}

	// Create and send on encoder
	enc := gob.NewEncoder(network)
	err := enc.Encode(jsonPkg)
	if err != nil {
		log.Panic("Encode error, ", err)
	}
}

func recvPkg(network *bytes.Buffer) ExtPkg {
	var jsonPkg []byte

	dec := gob.NewDecoder(network)
	err := dec.Decode(&jsonPkg)
	if err != nil {
		log.Panic("Decode error, ", err)
	}

	var rawPkg ExtPkg

	err = json.Unmarshal(jsonPkg, &rawPkg)
	if err != nil {
		log.Panic("Unmarshal error, ", err)
	}

	return rawPkg
}

func TestDecoder(t *testing.T) {
	ioman := NewIOManager()
	go ioman.Run()
	var network bytes.Buffer
	dec := NewJSONDecoder(&network)
	go ioman.Listen(dec)
}
