package main

import (
	"github.com/maxnordlund/breamio/aioli"
	"github.com/maxnordlund/breamio/aioli/client"

	"container/ring"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
)

type Registrater interface {
	Register(Key, chan<- Key) error
	Unregister(Key)
	io.Closer
}

var DoubleBinding = errors.New("Key combination already binded.")

type Action struct {
	Id    int
	Event string
	Data  interface{}
}

var (
	r       Registrater
	c       *client.Client
	logger  *log.Logger
	actions map[Key]*ring.Ring = make(map[Key]*ring.Ring)
)

func init() {
	runtime.GOMAXPROCS(1)
	logger = log.New(os.Stderr, "[Hotkeys] ", log.LstdFlags)
	conn, err := net.Dial("tcp", "localhost:4041")
	if err != nil {
		logger.Fatalln("Unable to establish connection to main application.", err)
	}
	c = client.NewClient(conn)
	logger.Println("Connected to main application.")
}

func main() {
	shutdown := make(chan os.Signal)
	signal.Notify(shutdown, os.Interrupt)

	presses := make(chan Key, 1)
	defer close(presses) //All users of channel should be done when this runs.

	defer func() {
		for key := range actions {
			r.Unregister(key)
		}
	}()

	key := Key{Alt, 'B'}
	actions[key] = ring.New(2)
	actions[key].Value = Action{1, "drawer:pause", struct{}{}}
	actions[key].Next().Value = Action{1, "drawer:resume", struct{}{}}

	err := r.Register(key, presses)
	if err != nil {
		logger.Println(err)
	}

	for {
		select {
		case key := <-presses:
			c.Send(toExtPkg(actions[key].Value.(Action)))
			actions[key] = actions[key].Next()
		case <-shutdown:
			return
		}
	}

	logger.Println("Over and out!")
}

func toExtPkg(a Action) aioli.ExtPkg {
	data, err := json.Marshal(a.Data)
	if err != nil {
		panic(err)
	}
	return aioli.ExtPkg{
		Event: a.Event,
		ID:    a.Id,
		Data:  data,
	}
}
