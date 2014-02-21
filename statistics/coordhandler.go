package main

import (
	"time"
)

//A Coordinate handler is a module that can receive Coordinate structs from a coordSource and maintain them in chronological order. Structs older then now-interval is automagically discarded.
type CoordinateHandler interface {
	GetCoords() (coords chan *Coordinate) //Returns a channel containing all cordinates in the CordinateHandler sorted chronologically
	SetInterval(interval time.Duration)
	SetDesiredFreq(desiredFreq int)
}

//listenTo is the channel that the coordinateHandler is should listen to
//interval is how old data we accept in this timelist
//desiredFreq is an upper limit on data per second that the coordinatehandler accepts
func NewCoordinateHandler(coordSource <-chan *Coordinate, interval time.Duration, desiredFreq int) *TimeList {
	return newTimeList(coordSource, interval, desiredFreq)
}

//A coordinate represents a point on the screen at a certain time
type Coordinate struct {
	x, y      float64
	timestamp time.Time
}

//Create a new Coordinate struct
//x and y is the coordinates that should be represented and timestamp is the time when the coordinate current

func NewCoordinate(x, y float64, timestamp time.Time) *Coordinate {
	return &Coordinate{
		x:         x,
		y:         y,
		timestamp: timestamp,
	}
}
