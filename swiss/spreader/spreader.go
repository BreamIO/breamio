package spreader

import (
	"encoding/json"
	"github.com/maxnordlund/breamio/aioli"
	"github.com/maxnordlund/breamio/aioli/client"
	bl "github.com/maxnordlund/breamio/beenleigh"
	"github.com/maxnordlund/breamio/briee"
	"github.com/maxnordlund/breamio/gorgonzola"
	"log"
	"net"
	"os"
	"flag"
	"fmt"
)

var mainServerAddr = flag.String("ip", "localhost", "This is the ip that the main server is located on.")
var mainServerPort = flag.Int("port", 4041, "This is the port that the main server is listening to.")

func init() {
	bl.Register(bl.NewRunHandler(startup))
}

func startup(logic bl.Logic, closer <-chan struct{}) {
	MainServer := fmt.Sprintf("%s:%ds",mainServerAddr, mainServerPort)
	logger := log.New(os.Stdout, "[Distributor] ", log.LstdFlags)
	conn, err := net.Dial("tcp", MainServer)
	defer func() {
		conn.Close()
		logger.Println("Shutting down.")
	}()

	if err != nil {
		logger.Printf("Unable to dial %s.", MainServer)
		return
	}

	c := client.NewClient(conn)
	newListenerChan := logic.RootEmitter().Subscribe("new:etlistener", bl.Spec{}).(<-chan bl.Spec)

	for {
		defer logic.RootEmitter().Unsubscribe("new:etlistener", newListenerChan)
		select {
		case <-closer:
			return
		case event := <-newListenerChan:
			l := newListener(event.Emitter, logic.CreateEmitter(event.Emitter), c, closer)

		}
	}
}

type proxy struct {
	subs     briee.Subscriber
	closer   <-chan struct{}
	dataChan <-chan *gorgonzola.ETData
}

func newListener(id int, subs briee.Subscriber, c *client.Client, closer <-chan struct{}) *proxy {
	listener := &proxy{
		closer:   closer,
		subs:     subs,
		dataChan: subs.Subscribe("tracker:etdata", &gorgonzola.ETData{}).(<-chan *gorgonzola.ETData),
	}
	go func() {
		defer listener.subs.Unsubscribe("tracker:etdata", listener.dataChan)
		for {
			select {
			case <-listener.closer:
				return
			case data := <-listener.dataChan:
				dat, err := json.Marshal(data)
				if err != nil {
					fmt.Println(err)
				} else {
					c.Send(aioli.ExtPkg{
						Event: "tracker:etdata",
						ID:    id,
						Data:  dat,
					})
				}
			}
		}
	}()
	return listener
}
