package ancientPower

import (
	"encoding/binary"
	bl "github.com/maxnordlund/breamio/beenleigh"
	"github.com/maxnordlund/breamio/briee"
	"github.com/maxnordlund/breamio/gorgonzola"
	"io"
	"log"
	"math"
	"net"
	"os"
)

var logger = log.New(os.Stdout, "[AncientPower] ", log.LstdFlags)

func New(logic bl.Logic, spec bl.Spec) {
	ee := logic.CreateEmitter(spec.Emitter)
	//addr := ":303" + strconv.Itoa(spec.Emitter)
	addr := spec.Data
	go ListenAndServe(ee, byte(spec.Emitter), addr)
	logger.Printf("Created a new AncientPower Server address %s on EE %d.\n", addr, spec.Emitter)
	return
}

func ListenAndServe(ee briee.EventEmitter, id byte, addr string) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Printf("Error attempting to listen to %s: %s", addr, err)
		return
	}
	logger.Printf("Listening on behalf of Event Emitter %d on %s.", id, addr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			logger.Println("Error accepting client: %s", err)
			return
		}
		logger.Printf("Received connection from %s.", conn.RemoteAddr())
		go (&client{conn, ee, false, id}).handle()
	}
}

var bufferQueue = make(chan []byte, 10)

func take() (buffer []byte) {
	select {
	case buffer = <-bufferQueue:
	default:
		buffer = make([]byte, 64) //Make new one.
	}
	return
}

func giveBack(buffer []byte) {
	select {
	case bufferQueue <- buffer[:cap(buffer)]: //Return for recycling.
	default: // Drop it.
	}
}

func (c *client) handle() {
	defer c.Close() //Make sure connection is closed before leaving.

	go func() {
		etCh := c.ee.Subscribe("tracker:etdata", &gorgonzola.ETData{}).(<-chan *gorgonzola.ETData)
		defer c.ee.Unsubscribe("tracker:etdata", etCh)
		for data := range etCh {
			//logger.Println("Recieved on tracker:etdata :", data)
			if c.subscribing {
				buffer := take()
				buffer[0] = 1
				binary.BigEndian.PutUint64(buffer[1:9], math.Float64bits(data.Filtered.X()))
				binary.BigEndian.PutUint64(buffer[9:17], math.Float64bits(data.Filtered.Y()))
				binary.BigEndian.PutUint64(buffer[17:25], uint64(data.Timestamp.Unix()))
				if _, err := c.Write(buffer[:25]); err != nil {
					return
				}
			}
		}
		c.Close() // Basically signal client that it is done.
	}()
	for {
		buffer := take()
		defer giveBack(buffer)

		//Use buffer
		_, err := c.Read(buffer[:1])
		if err != nil {
			return
		}
		switch buffer[0] {
		case 1:
			c.getPoints() //Request ETData
		case 7:
			c.name() //Name
		case 8:
			c.fps() //FPS
		case 9:
			c.keepalive() //KeepAlive
		default:
			return //Invalid package. Drop client.
		}
	}
}

// 1
func (c *client) getPoints() {
	logger.Println("1")
	c.subscribing = !c.subscribing
	buffer := take()[:25]
	defer giveBack(buffer)
	c.Read(buffer[1:25])
}

// 7
func (c *client) name() {
	logger.Println("7")
	buffer := take()[:1]
	defer giveBack(buffer)
	c.Read(buffer)
	c.Write([]byte{7, c.id})
}

// 8
func (c *client) fps() {
	logger.Println("8")
	buffer := take()[:1]
	defer giveBack(buffer)
	c.Read(buffer)
	c.Write([]byte{8, 40}) //Screw this guys, it is not like if anyone cares anyway.
}

// 9
func (c *client) keepalive() {
	logger.Println("9")
	c.Write([]byte{9})
}

type client struct {
	io.ReadWriteCloser
	ee          briee.EventEmitter
	subscribing bool
	id          byte
}
