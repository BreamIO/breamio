package aioli

import (
	"bytes"
	"encoding/json"
	"log"
	"sync"
	//"time"
	"github.com/maxnordlund/breamio/briee"
	"io"
	"testing"
	"os"
	//"reflect"
)

type Payload struct {
	X, Y float64
}

func send(pl Payload, network io.Writer) {
	// Encode Data, representing the other side, e.g. WEB, CLI
	// Encode the payload data of type Payload as json
	jsonpl, err := json.Marshal(pl)
	if err != nil {
		log.Panic("Marshal error, ", err)
	}

	rawPkg := ExtPkg{
		Event: "event data",
		ID:    1,
		Data:  jsonpl,
	}

	enc := json.NewEncoder(network)
	err = enc.Encode(rawPkg)
	if err != nil {
		log.Panic("Encode error, ", err)
	}
}

func recvPkg(network io.Reader) ExtPkg {
	var rawPkg ExtPkg

	dec := json.NewDecoder(network)
	dec.Decode(&rawPkg)

	return rawPkg
}

func TestIOman(t *testing.T) {
	// Set up emitter
	ee := briee.New()
	//log.Println("[EE Run]")
	go ee.Run()
	subscriber := ee.Subscribe("event data", Payload{}).(<-chan Payload)

	// Set up IO manager
	ioman := New()

	// Add event emitter
	err := ioman.AddEE(ee, 1)
	if err != nil {
		t.Errorf("Unable to add event emitter")
	}

	var network bytes.Buffer // Stand-in for the network
	dec := json.NewDecoder(&network)

	// Example data from an external source
	plSend := Payload{
		X: 0.1,
		Y: 0.2,
	}

	var plRecv Payload

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		// Send decodes and sends payload data on network
		send(plSend, &network)
		//log.Println("[Send]")
		//send(plSend, network)
		wg.Done()
	}()

	go func() {
		// Listen on subscriber channel
		// plRecv = <-subscriber
		plRecv = <-subscriber
		//log.Println("[Receive]")
		wg.Done()
	}()

	go func(){
		//log.Println("[Manager Run]")
		ioman.Run()
	}()

	go func(){
		//log.Println("[Manager Listen]")
		logger := log.New(os.Stdout, "[AIOLI Test]", log.LstdFlags)
		ioman.Listen(dec, logger)
	}()

	wg.Wait()
	if plSend != plRecv {
		t.Errorf("Got %v, want %v\n", plRecv, plSend)
	}

	ioman.Close()
}

func TestAddRemoveEmitters(t *testing.T) {
	ee := briee.New()
	ioman := New()

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
	ee := briee.New()
	ioman := New()

	// Add event emitter
	err := ioman.AddEE(ee, 0)
	if err == nil {
		t.Errorf("Should not be able to add broadcast identifier")
	}
}

func TestRemEEBC(t *testing.T) {
	ee := briee.New()
	ioman := New()

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


func TestDecoder(t *testing.T) {
	ioman := New()
	go ioman.Run()
	var network bytes.Buffer
	dec := NewJSONDecoder(&network)
	logger := log.New(os.Stdout, "[AIOLI Test]", log.LstdFlags)
	go ioman.Listen(dec, logger)
}

