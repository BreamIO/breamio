package gorgonzola

import (
	"github.com/maxnordlund/breamio/briee"
)

func Link(ee briee.EventEmitter, t Tracker) error {
	publisher := ee.Publish("tracker:etdata", &ETData{}).(chan<- *ETData)

	dataCh, errCh := t.Stream()
	defer func() {
		if r := recover(); r != nil {
			println("Caught a runtime panic:", r)
			//Recover from a close on the publisher channel.
			//Do not want to bring down entire application
		}
	}()

	return listenAndServe(dataCh, errCh, publisher)
}

func listenAndServe(dataChannel <-chan *ETData, errorChannel <-chan error, publisher chan<- *ETData) error {
	for {
		select {
		case data, ok := <-dataChannel:
			if !ok {
				return nil //No more data from tracker. Shutting down. Does not break due to weirdness in testing.
			}
			select {
			case publisher <- data: // Attempt to send
			default:
				println("[Gorgonzola] Dropped package due to full channel.")
			}
		case err := <-errorChannel:
			return err
		}
	}
	return nil //Dead code, but compiler insists on its existence
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
	return point2D{(left.X() + right.X()) / 2, (left.Y() + right.Y()) / 2}
}
