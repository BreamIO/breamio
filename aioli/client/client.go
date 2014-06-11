package client

import (
	"github.com/maxnordlund/breamio/aioli"
	//	"github.com/maxnordlund/breamio/beenleigh"
	"io"
	"log"
	"sync"
	//	"net"
)

/*
func main() {
	conn, err := net.Dial("tcp", "localhost:4041")
	defer conn.Close()
	if err != nil {
		log.Println("Could not connnect to server:", err)
		return
	}
	c := NewClient(conn)
	payload, err := json.Marshal(beenleigh.Spec{1, "mock://standard"})
	c.Send(aioli.ExtPkg{"new:tracker", 256, payload})
	c.Wait()
}*/

type Client struct {
	out    chan aioli.ExtPkg
	in     chan aioli.ExtPkg
	encdec aioli.EncodeDecoder
	wg     sync.WaitGroup
	io.ReadWriteCloser
}

func NewClient(conn io.ReadWriteCloser) *Client {
	out := make(chan aioli.ExtPkg)
	in := make(chan aioli.ExtPkg)
	c := &Client{out, in, aioli.NewCodec(conn), sync.WaitGroup{}, conn}
	go c.run()
	return c
}

func (c *Client) run() {
	defer c.Close()

	go func() {
		for {
			var pkg aioli.ExtPkg
			err := c.encdec.Decode(pkg)
			if err != nil {
				return
			}
			c.in <- pkg
		}
	}()

	for pkg := range c.out {
		if err := c.encdec.Encode(pkg); err != nil {
			log.Println("Error writing package to Writer:", err)
			return
		}
		c.wg.Done()
	}
}

func (c *Client) Send(pkg aioli.ExtPkg) {
	c.wg.Add(1)
	c.out <- pkg
}

func (c *Client) Recieve() (pkg aioli.ExtPkg) {
	return <-c.in
}

func (c *Client) Wait() {
	c.wg.Wait()
}
