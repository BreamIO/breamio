package analysis

import (
	"time"

	gr "github.com/maxnordlund/breamio/gorgonzola"
)

//A ETData handler is a module that can receive ETData structs from a coordSource and maintain them in chronological order. Structs older then now-interval is automagically discarded.
type CoordinateHandler interface {
	GetCoords() (coords chan *gr.ETData) //Returns a channel containing all cordinates in the CordinateHandler sorted chronologically
	SetInterval(interval time.Duration)
	SetDesiredFreq(desiredFreq uint)
	Flush() //Flushes data (coords)
	Start() //Makes sure the buffer is running but flushes no data(cords)
	Stop()  //Makes sure the buffer is stopped and flushes data(coords)
}
