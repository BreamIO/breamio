package statistics

import (
	"time"
)

type CoordBuffer struct {
	interval    time.Duration
	desiredFreq int
	data        []ETData
	start, end  int
}

// Create a new CoordBuffer
// it implements the CoordinateHandler interface
func newCoordBuffer(coordSource <-chan *ETData, interval time.Duration, desiredFreq int) *CoordBuffer {
	//TODO  start a go routine that adds coords from coordsource
	c := &CoordBuffer{
		interval:    interval,
		desiredFreq: desiredFreq,
		// One extra data to allow almost overlapping
		data:  make([]ETData, desiredFreq*int(interval.Seconds())+1),
		start: 0,
		end:   0, //End is not included in the list
	}

	go func() {
		for d := range coordSource {
			c.add(d)
		}
	}()

	return c
}

// Returns a channel containing all ETData structs in
// t sorted chronologically
func (c *CoordBuffer) GetCoords() (coords chan *ETData) {
	coords = make(chan *ETData)

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

func (c *CoordBuffer) add(coord *ETData) {
	c.data[c.end] = *coord

	c.end = (c.end + 1) % len(c.data)

	// if c.end == c.start {
	// 	c.start = (c.start + 1) % len(c.data)
	// }
}

// Used to make sure the data you get is always fresh.
func (c *CoordBuffer) refresh() {
	for time.Since(c.data[c.start].Timestamp) > c.interval {
		c.start = (c.start + 1) % len(c.data)
		if c.start == c.end {
			break
		}
	}
}

//Currently removes all data collected if duration updates
func (c *CoordBuffer) SetInterval(interval time.Duration) {
	c.interval = interval
	c.data = make([]ETData, c.desiredFreq*int(c.interval.Seconds()))
}

//Currently removes all data if desiredFreq updates
func (c *CoordBuffer) SetDesiredFreq(desiredFreq int) {
	c.desiredFreq = desiredFreq
	c.data = make([]ETData, c.desiredFreq*int(c.interval.Seconds()))
}
