package statistics

import (
	"time"
)

//A ETData handler is a module that can receive ETData structs from a coordSource and maintain them in chronological order. Structs older then now-interval is automagically discarded.
type CoordinateHandler interface {
	GetCoords() (coords chan *ETData) //Returns a channel containing all cordinates in the CordinateHandler sorted chronologically
	SetInterval(interval time.Duration)
	SetDesiredFreq(desiredFreq int)
}

//listenTo is the channel that the coordinateHandler is should listen to
//interval is how old data we accept in this timelist
//desiredFreq is an upper limit on data per second that the coordinatehandler accepts
func NewCoordinateHandler(coordSource <-chan *ETData, interval time.Duration, desiredFreq int) *CoordBuffer {
	return newCoordBuffer(coordSource, interval, desiredFreq)
}
