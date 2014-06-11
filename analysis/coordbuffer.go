package analysis

import (
	"time"

	gr "github.com/maxnordlund/breamio/gorgonzola"
)

type CoordBuffer struct {
	interval    time.Duration
	desiredFreq uint
	data        []gr.ETData
	start, end  int
	running     bool
}

// Create a new CoordBuffer
// it implements the CoordinateHandler interface
func NewCoordBuffer(coordSource <-chan *gr.ETData, interval time.Duration, desiredFreq uint) *CoordBuffer {
	c := &CoordBuffer{
		interval:    interval,
		desiredFreq: desiredFreq,
		// One extra data to allow almost overlapping
		data:    make([]gr.ETData, desiredFreq*uint(interval.Seconds())+1),
		start:   0,
		end:     0, //End is not included in the list
		running: true,
	}

	go func() {
		for d := range coordSource {
			if c.running {
				c.add(d)
			}
		}
	}()

	return c
}

// Returns a channel containing all ETData structs in
// t sorted chronologically
func (c *CoordBuffer) GetCoords() (coords chan *gr.ETData) {
	coords = make(chan *gr.ETData)
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

func (c *CoordBuffer) add(coord *gr.ETData) {
	c.data[c.end] = *coord

	c.end = (c.end + 1) % len(c.data)

	if c.end == c.start {
		c.start = (c.start + 1) % len(c.data)
	}
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
	c.data = make([]gr.ETData, c.desiredFreq*uint(c.interval.Seconds())+1)
}

//Currently removes all data if desiredFreq updates
func (c *CoordBuffer) SetDesiredFreq(desiredFreq uint) {
	c.desiredFreq = desiredFreq
	c.data = make([]gr.ETData, c.desiredFreq*uint(c.interval.Seconds())+1)
}

//Empties the buffer. That is removing all cords.
func (c *CoordBuffer) Flush() {
	c.end = c.start
}

//Starts the buffer if it is stopped. Does nothing if it is running.
func (c *CoordBuffer) Start() {
	c.running = true
}

//Stops the buffer and flushes the data if the buffer is running. Else it does nothing.
func (c *CoordBuffer) Stop() {
	if c.running {
		c.running = false
		c.Flush()
	}
}
