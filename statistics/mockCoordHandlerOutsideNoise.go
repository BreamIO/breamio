package main

import (
	"time"
	"math/rand"
)

//This is a mockCoordHandler made for testing
//It will not care for anything you tell it except desiredFreq
//The only thing it does is sending random nois outside of the screen when asking for coordinates.
//It will give you desiredFreq*80 points

type MockCoordHandlerOutsideNoise struct {
	desiredFreq int
}

func NewMockCoordHandlerOutsideNoise(coordSource chan Coordinate, interval time.Duration, desiredFreq int) *MockCoordHandlerOutsideNoise {
	return &MockCoordHandlerOutsideNoise {
		desiredFreq: desiredFreq,
	}
}


func (m MockCoordHandlerOutsideNoise) GetCoords() (coords chan *Coordinate) {
	coords = make(chan *Coordinate)

	for i := 0; i<m.desiredFreq*10; i++ {
		coords <- NewCoordinate(0 - rand.Float64(), rand.Float64(), time.Now())
		coords <- NewCoordinate(1 + rand.Float64(), rand.Float64(), time.Now())
		coords <- NewCoordinate(rand.Float64(), 0 - rand.Float64(), time.Now())
		coords <- NewCoordinate(rand.Float64(), 1 + rand.Float64(), time.Now())
		coords <- NewCoordinate(0 - rand.Float64(), 0 - rand.Float64(), time.Now())
		coords <- NewCoordinate(1 + rand.Float64(), 0 - rand.Float64(), time.Now())
		coords <- NewCoordinate(0 - rand.Float64(), 1 + rand.Float64(), time.Now())
		coords <- NewCoordinate(1 + rand.Float64(), 1 + rand.Float64(), time.Now())
	}
	return coords
}
