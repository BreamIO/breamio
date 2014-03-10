package statistics

/*
import (
	"time"
)

//This is a mockCoordHandler made for testing
//It will not care for anything you tell it except for desired freq
//The only thing it does is generate coordinates that first looks across diagonals of the screen (first \ then /) and then horizontally in the middle(like -) and finally vertically in the middle(like |)
//It prioritizes looking right to left and then top to bottom
//The lines will consist of desiredFreq points each.

type MockCoordHandlerCrosses struct {
	desiredFreq int
	scaling     float64
}

func NewMockCoordHandlerCrosses(coordSource chan Coordinate, interval time.Duration, desiredFreq int) *MockCoordHandlerCrosses {
	return &MockCoordHandlerCrosses{
		//Dump everything except desiredFreq, we don't care for parameters since we want a controlled testing environment.
		desiredFreq: desiredFreq,
		scaling:     float64(desiredFreq),
	}
}

func (m MockCoordHandlerCrosses) GetCoords() (coords chan *Coordinate) {
	coords = make(chan *Coordinate)

	//Generate coords as \
	for i := 0; i < m.desiredFreq; i++ {
		coords <- NewCoordinate(float64(i)/m.scaling, float64(i)/m.scaling, time.Now())
	}

	//Generate coords as /
	for i := 0; i < m.desiredFreq; i++ {
		coords <- NewCoordinate(1-(float64(i)/m.scaling), float64(i)/m.scaling, time.Now())
	}

	//Generate coords as -
	for i := 0; i < m.desiredFreq; i++ {
		coords <- NewCoordinate(float64(i)/m.scaling, 0.5, time.Now())
	}

	//Generate coords as |
	for i := 0; i < m.desiredFreq; i++ {
		coords <- NewCoordinate(0.5, float64(i)/m.scaling, time.Now())
	}
	return coords
}
*/
