package main

import (
	"github.com/maxnordlund/breamio/aioli"
	"github.com/maxnordlund/breamio/briee"
	bl "github.com/maxnordlund/breamio/beenleigh"
)

func main() {
	ee := briee.New()
	io := aioli.New()
	logic := bl.New(ee, io)
	logic.ListenAndServe()
}