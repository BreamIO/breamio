package gorgonzola

import (
	"math"
	"time"
)

type MockTracker struct {
	f func() (float64, float64)
}
	
func (m MockTracker) Stream() (<-chan *ETData, <-chan error) {
	ch := make(chan *ETData)
	errs := make(chan error, 1)
	go func(){
		if m.f == nil {
			return
		}
		x, y := m.f()
		ch<-&ETData{Point2D{x, y}, time.Now()}
		
	}()
	return ch, errs
}

func (m *MockTracker) Close() error {
	m.f = nil
	return nil
}

func (m *MockTracker) Connect() error {
	t := float64(0)
	m.f = func() (float64, float64) {
		defer func(){t+=.1}()
		return math.Cos(t), math.Sin(t)
	}
	return nil
}

func (m *MockTracker) Calibrate(points <-chan Point2D, errs chan<- error) {
	errs <- NotImplementedError("Calibrate of MockTracker")
}

func (m *MockTracker) IsCalibrated() bool {
	return true
}

