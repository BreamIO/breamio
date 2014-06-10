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
)

const (
	MainServer = "localhost:4041"
)

func init() {
	bl.Register(bl.NewRunHandler(startup))
}

func startup(logic bl.Logic, closer <-chan struct{}) {
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
		defer listener.ee.Unsubscribe("new:etlistener", newListenerChan)
		select {
		case <-closer:
			return
		case event := <-newListenerChan:
			l := newListener(logic.CreateEmitter(event.Emitter), c, closer)

		}
	}
}

type listener struct {
	subs     briee.Subscriber
	closer   <-chan struct{}
	dataChan <-chan *gorgonzola.ETData
}

func newListener(id int, subs briee.Subscriber, c *client.Client, closer <-chan struct{}) *listener {
	listener := &listener{
		closer:   closer,
		subs:     subs,
		dataChan: subs.Subscribe("tracker:etdata", &gorgonzola.ETData{}).(<-chan *gorgonzola.ETData),
	}
	go func() {
		defer listener.ee.Unsubscribe("tracker:etdata", listener.dataChan)
		for {
			select {
			case <-listener.closer:
				return
			case data := <-listener.dataChan:
				c.Send(aioli.ExtPkg{
					Event: "tracker:etdata",
					ID:    id,
					Data:  json.Marshal(data),
				})
			}
		}
	}()
}
