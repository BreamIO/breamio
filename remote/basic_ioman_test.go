package remote_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"os"
	"runtime/pprof"
	"sync"
	"testing"

	"github.com/maxnordlund/breamio/briee"
	bl "github.com/maxnordlund/breamio/moduler"
	. "github.com/maxnordlund/breamio/remote"
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
	logic := bl.New(briee.New)
	// Set up IO manager
	ioman := New(logic)
	defer ioman.Close()

	// Set up emitter
	ee := logic.CreateEmitter(1)
	subscriber := ee.Subscribe("event data", Payload{}).(<-chan Payload)
	defer ee.Unsubscribe("event data", subscriber)

	network := SyncReadWriter{RW: &bytes.Buffer{}} // Stand-in for the network
	dec := NewCodec(&network)

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

	go func() {
		//log.Println("[Manager Run]")
		ioman.Run()
	}()

	go func() {
		//log.Println("[Manager Listen]")
		logger := log.New(os.Stdout, "[AIOLI Test]", log.LstdFlags)
		ioman.Listen(dec, logger)
	}()

	wg.Wait()
	if plSend != plRecv {
		t.Errorf("Got %v, want %v\n", plRecv, plSend)
	}
}

func TestSubscriptions(t *testing.T) {
	logic := bl.New(briee.New)
	// Set up IO manager
	ioman := New(logic)
	defer ioman.Close()

	ee := logic.CreateEmitter(1)
	pub := ee.Publish("data", string("")).(chan<- string)

	buffer := &bytes.Buffer{}
	logger := log.New(buffer, "[AIOLI Test]", log.LstdFlags)

	go ioman.Run()

	barrier := make(chan struct{})
	network := &SyncReadWriter{RW: &bytes.Buffer{}} // Stand-in for the network

	go func() {
		serverCodec := NewCodec(network)
		go ioman.Listen(serverCodec, logger)
		<-barrier
		pub <- "Foo"
	}()

	clientCodec := NewCodec(network)

	//Write various packages to network.
	clientCodec.Encode(ExtPkg{
		Event:     "data",
		Subscribe: true,
		ID:        1,
		Data:      []byte(""),
	})

	barrier <- struct{}{}

	var ans struct{ S string }
	clientCodec.Decode(&ans)
	t.SkipNow()
	if ans.S != "Foo" {
		t.Errorf("Wrong data in answer. Expected \"Foo\", found \"%s\".", ans.S)
	}
}
