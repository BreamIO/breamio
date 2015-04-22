package proxy

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/maxnordlund/breamio/briee"
	gr "github.com/maxnordlund/breamio/eyetracker"
	bl "github.com/maxnordlund/breamio/moduler"
	"github.com/maxnordlund/breamio/remote"
	"github.com/maxnordlund/breamio/remote/access"
	"github.com/maxnordlund/breamio/remote/client"
	"log"
	"net"
	"os"
)

var mainServerAddr = flag.String("ip", "localhost", "This is the ip that the main server is located on.")
var mainServerPort = flag.Int("port", 4041, "This is the port that the main server is listening to.")

func init() {
	bl.Register(bl.NewRunHandler(startup))
}

func startup(logic bl.Logic, closer <-chan struct{}) {
	MainServer := fmt.Sprintf("%s:%d", *mainServerAddr, *mainServerPort)
	logger := log.New(os.Stdout, "[Distributor] ", log.LstdFlags)
	conn, err := net.Dial("tcp", MainServer)

	if err != nil {
		logger.Printf("Unable to dial %s. Shutting down.", MainServer)
		os.Exit(1)
	}

	defer func() {
		conn.Close()
		logger.Println("Shutting down.")
	}()

	c := client.NewClient(conn)
	
	//TODO remove todo
	
	//TODO Add close chan for routine (not relevant on DH14)
	go func() { //Listen for all external events
		ioman := access.GetIOManager()
		for {
			pkg := c.Recieve()
			ioman.Dispatch(pkg)
		}
	}()

	newListenerChan := logic.RootEmitter().Subscribe("new:etlistener", bl.Spec{}).(<-chan bl.Spec)
	for {
		defer logic.RootEmitter().Unsubscribe("new:etlistener", newListenerChan)
		select {
		case <-closer:
			logger.Printf("Shutting down")
			return
		case event := <-newListenerChan:
			logger.Printf("Starting a listener to emitter:", event.Emitter)
			newListener(event.Emitter, logic.CreateEmitter(event.Emitter), c, closer)
		}
	}
}

type etDataProxy struct {
	subs        briee.Subscriber
	closer      <-chan struct{}
	dataChan    <-chan *gr.ETData
	calNextChan <-chan struct{}
	calEndChan  <-chan struct{}
	valNextChan <-chan struct{}
	valEndChan  <-chan float64
}

func newListener(id int, subs briee.Subscriber, c *client.Client, closer <-chan struct{}) *etDataProxy {
	listener := &etDataProxy{
		closer:      closer,
		subs:        subs,
		dataChan:    subs.Subscribe("tracker:etdata", &gr.ETData{}).(<-chan *gr.ETData),
		calNextChan: subs.Subscribe("tracker:calibrate:next", struct{}{}).(<-chan struct{}),
		calEndChan:  subs.Subscribe("tracker:calibrate:end", struct{}{}).(<-chan struct{}),
		valNextChan: subs.Subscribe("tracker:validate:next", struct{}{}).(<-chan struct{}),
		valEndChan:  subs.Subscribe("tracker:validate:end", float64(0)).(<-chan float64),
	}
	go func() {
		defer listener.subs.Unsubscribe("tracker:etdata", listener.dataChan)
		defer listener.subs.Unsubscribe("tracker:calibrate:next", listener.calNextChan)
		defer listener.subs.Unsubscribe("tracker:calibrate:end", listener.calEndChan)
		defer listener.subs.Unsubscribe("tracker:validate:next", listener.valNextChan)
		defer listener.subs.Unsubscribe("tracker:validate:end", listener.valEndChan)
		for {
			select {
			case <-listener.closer:
				return
			case data := <-listener.dataChan:
				proxySend(c, "tracker:etdata", id, data)
			case data := <-listener.calNextChan:
				proxySend(c, "tracker:calibrate:next", id, data)
			case data := <-listener.calEndChan:
				proxySend(c, "tracker:calibrate:end", id, data)
			case data := <-listener.valNextChan:
				proxySend(c, "tracker:validate:next", id, data)
			case data := <-listener.valEndChan:
				proxySend(c, "tracker:validate:end", id, data)
			}
		}
	}()
	return listener
}

func proxySend(c *client.Client, event string, id int, data interface{}) {
	dat, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
	} else {
		c.Send(remote.ExtPkg{
			Event: event,
			ID:    id,
			Data:  dat,
		})
	}
}
