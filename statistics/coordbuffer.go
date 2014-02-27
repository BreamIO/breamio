package main

import (
	"time"
)

type CoordBuffer struct {
	interval    time.Duration
	desiredFreq int
	data        []Coordinate
	start, end  int
}

/*
Create a new CoordBuffer
it implements the CoordinateHandler interface
*/
func NewCoordBuffer(coordSource <-chan *Coordinate, interval time.Duration, desiredFreq int) *CoordBuffer {
	//TODO  start a go routine that adds coords from coordsource
	return &CoordBuffer{
		interval:    interval,
		desiredFreq: desiredFreq,
		data:        make([]Coordinate, desiredFreq*int(interval.Seconds())),
		start:       0,
		end:         0, //End is not included in the list
	}
}

/*
Returns a channel containing all Coordinate structs in t sorted chronologically
*/
func (c CoordBuffer) GetCoords() (coords chan *Coordinate) {
	coords = make(chan *Coordinate)

	c.refresh()

	go func() {
		for i := c.start; i != c.end; {
			coords <- &c.data[i]
			i = (i + 1) % len(c.data)
		}
		close(coords)
	}()

	return coords
}

func (c CoordBuffer) add(coord *Coordinate) {
	c.data[c.end] = *coord

	if c.end == c.start {
		c.start = (c.start + 1) % len(c.data)
	}
}

/*
func (c CoordBuffer) getAddFunc() func(*Coordinate) {
	return func(coord *Coordinate) {
		c.Add(coord)
	}
}
*/

/*Used to make sure the data you get is always fresh.*/
func (c CoordBuffer) refresh() {
	for time.Since(c.data[c.start].timestamp) > c.interval {
		c.start = (c.start + 1) % len(c.data)
		if c.start == c.end {
			break
		}
	}
}

//Currently removes all data collected if duration updates
func (c CoordBuffer) SetInterval(interval time.Duration) {
	c.interval = interval
	c.data = make([]Coordinate, c.desiredFreq*int(c.interval.Seconds()))
}

//Currently removes all data if desiredFreq updates
func (c CoordBuffer) SetDesiredFreq(desiredFreq int) {
	c.desiredFreq = desiredFreq
	c.data = make([]Coordinate, c.desiredFreq*int(c.interval.Seconds()))
}
