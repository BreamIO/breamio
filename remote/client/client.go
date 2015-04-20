package client

import (
	"github.com/maxnordlund/breamio/remote"
	"io"
	"log"
	"sync"
)

type Client struct {
	out    chan remote.ExtPkg
	in     chan remote.ExtPkg
	encdec remote.EncodeDecoder
	wg     sync.WaitGroup
	io.ReadWriteCloser
}

func NewClient(conn io.ReadWriteCloser) *Client {
	out := make(chan remote.ExtPkg)
	in := make(chan remote.ExtPkg)
	c := &Client{out, in, remote.NewCodec(conn), sync.WaitGroup{}, conn}
	go c.run()
	return c
}

func (c *Client) run() {
	defer c.Close()

	go func() {
		for {
			var pkg remote.ExtPkg
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

func (c *Client) Send(pkg remote.ExtPkg) {
	c.wg.Add(1)
	c.out <- pkg
}

func (c *Client) Recieve() (pkg remote.ExtPkg) {
	return <-c.in
}

func (c *Client) Wait() {
	c.wg.Wait()
}
