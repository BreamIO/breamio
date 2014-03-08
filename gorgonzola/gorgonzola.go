package gorgonzola

import (
	"github.com/maxnordlund/breamio/briee"
)

func Link(ee briee.EventEmitter, t Tracker) error {
	publisher := ee.Publish("tracker:etdata", &ETData{}).(chan<- *ETData)
	defer close(publisher)
	
	dataCh, errs := t.Stream()
	defer func() {
		if r := recover(); r != nil {
			println("Caught a runtime panic:", r)
			//Recover from a close on the publisher channel.
			//Do not want to bring down entire application
		}
	}()

	for {
		select {
		case data, ok := <-dataCh:
			if !ok {
				break //No more data from tracker. Shutting down.
			}
			select {
			case publisher <- data: // Attempt to send
			default:
				println("[Gorgonzola] Dropped package due to full channel.")
			}
		case err := <-errs:
			return err
		}
	}
	return nil
}

type Point2D interface {
	X() float64 
	Y() float64
}

type point2D struct {
	x, y float64 
}

func (p point2D) X() float64 {
	return p.x
}

func (p point2D) Y() float64 {
	return p.y
}

func filter(left, right Point2D) Point2D {
	return point2D{ (left.X() + right.X()) / 2, (left.Y() + right.Y()) / 2}
}
