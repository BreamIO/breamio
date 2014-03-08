package main

import (
	"fmt"
	bl "github.com/maxnordlund/breamio/beenleigh"
	"github.com/maxnordlund/breamio/briee"
	"github.com/maxnordlund/breamio/gorgonzola"
	
)

func main() {
	logic := bl.New()
	go etPrinter(logic.RootEmitter())
	newtrackerEvents := logic.RootEmitter().Publish("new:tracker", bl.TrackerSpec{}).(chan<- bl.TrackerSpec)
	newtrackerEvents <- bl.TrackerSpec{"mock", "constant", 256}
	logic.ListenAndServe()
}

func etPrinter(ee briee.EventEmitter) {
	etEvents := ee.Subscribe("tracker:etdata", &gorgonzola.ETData{}).(<-chan *gorgonzola.ETData)
	for event := range etEvents {
		fmt.Println(event)
	}
}