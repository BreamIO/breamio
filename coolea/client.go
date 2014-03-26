package main

import (
	"encoding/json"
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
	ch chan aioli.ExtPkg
	wg sync.WaitGroup
	io.WriteCloser
}

func NewClient(conn io.WriteCloser) *Client {
	ch := make(chan aioli.ExtPkg)
	c := &Client{ch, sync.WaitGroup{}, conn}
	go c.run()
	return c
}

func (c *Client) run() {
	defer c.Close()
	enc := json.NewEncoder(c)
	for pkg := range c.ch {
		if err := enc.Encode(pkg); err != nil {
			log.Println("Error writing package to Writer:", err)
		}
		c.wg.Done()
	}
}

func (c *Client) Send(pkg aioli.ExtPkg) {
	c.wg.Add(1)
	c.ch <- pkg
}

func (c *Client) Wait() {
	c.wg.Wait()
}
