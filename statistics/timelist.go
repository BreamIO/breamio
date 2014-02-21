package main

import (
	"time"
)

type TimeList struct {
	interval   time.Duration
	desiredFreq int
	data       []Coordinate
	start, end int
}

/*
Create a new timelist
it implements the CoordinateHandler interface
*/
func newTimeList(coordSource chan Coordinate, interval time.Duration, desiredFreq int) *TimeList {
	return &TimeList{
		interval: interval,
		desiredFreq: desiredFreq,
		data:     make([]Coordinate, desiredFreq*int(interval.Seconds())),
		start:    0,
		end:      0, //End is not included in the list
	}
}

/*
Returns a channel containing all Coordinate structs in t sorted chronologically
*/
func (t TimeList) GetCoords() (coords chan *Coordinate) {
	coords = make(chan *Coordinate)

	t.refresh()

	go func() {
		for i := t.start; i != t.end; {
			coords <- &t.data[i]
			i = (i + 1) % len(t.data)
		}
		close(coords)
	}()

	return coords
}

func (t TimeList) add(coord *Coordinate) {
	t.data[t.end] = *coord

	if t.end == t.start {
		t.start = (t.start + 1) % len(t.data)
	}
}

/*
func (t TimeList) getAddFunc() func(*Coordinate) {
	return func(coord *Coordinate) {
		t.Add(coord)
	}
}
*/

/*Used to make sure the data you get is always fresh.*/
func (t TimeList) refresh() {
	for time.Since(t.data[t.start].timestamp) > t.interval {
		t.start = (t.start + 1) % len(t.data)
		if t.start == t.end {
			break
		}
	}
}

//Currently removes all data collected if duration updates
func (t TimeList) SetInterval(interval time.Duration) {
	t.interval = interval
	t.data = make([]Coordinate, t.desiredFreq*int(t.interval.Seconds()))
}

//Currently removes all data if desiredFreq updates
func (t TimeList) SetDesiredFreq(desiredFreq int) {
	t.desiredFreq = desiredFreq
	t.data = make([]Coordinate, t.desiredFreq*int(t.interval.Seconds()))
}
