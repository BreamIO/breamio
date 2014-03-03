package gorgonzola

import (
	"github.com/maxnordlund/breamio/briee"
)

func Link(ee briee.EventEmitter, t Tracker) error {
	publisher := ee.Publish("tracker:etdata", &ETData{}).(chan<- *ETData)
	defer close(publisher)
	dataCh, err = t.Stream()
	if err != nil {
		return err
	}
	defer func() {
		if r:= recover(); r != nil {
			println("Caught a runtime panic:", r)
			//Recover from a close on the publisher channel.
			//Do not want to bring down entire application
		}
	}
	
	for data := range dataCh {
		select {
			publisher <- data
			default:
		}
	}
}
