package aioli_test

import (
	"bytes"
	"encoding/json"
	"log"
	"sync"
	//"time"
	"io"
	"net"
	"os"
	"testing"
	//"reflect"

	. "github.com/maxnordlund/breamio/aioli"
	bl "github.com/maxnordlund/breamio/beenleigh"
	"github.com/maxnordlund/breamio/briee"
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

	// Set up emitter
	ee := logic.CreateEmitter(1)
	subscriber := ee.Subscribe("event data", Payload{}).(<-chan Payload)
	defer ee.Unsubscribe("event data", subscriber)

	var network bytes.Buffer // Stand-in for the network
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

	ioman.Close()
}

func TestSubscriptions(t *testing.T) {
	logic := bl.New(briee.New)
	// Set up IO manager
	ioman := New(logic)

	ee := logic.CreateEmitter(1)
	pub := ee.Publish("data", string("")).(chan<- string)

	buffer := &bytes.Buffer{}
	logger := log.New(buffer, "[AIOLI Test]", log.LstdFlags)

	go ioman.Run()

	sync := make(chan struct{})
	server, err := net.Listen("tcp", "localhost:4042")
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		serverSocket, _ := server.Accept()
		serverCodec := NewCodec(serverSocket)
		go ioman.Listen(serverCodec, logger)
		<-sync
		pub <- "Foo"
	}()

	clientSocket, _ := net.Dial("tcp", "localhost:4042")
	clientCodec := NewCodec(clientSocket)

	//Write various packages to network.
	clientCodec.Encode(ExtPkg{
		Event:     "data",
		Subscribe: true,
		ID:        1,
		Data:      []byte(""),
	})

	sync <- struct{}{}

	var ans struct{ S string }
	clientCodec.Decode(&ans)
	t.SkipNow()
	if ans.S != "Foo" {
		t.Errorf("Wrong data in answer. Expected \"Foo\", found \"%s\".", ans.S)
	}
}
