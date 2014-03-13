package gorgonzola

import (
	"errors"
	"math"
	"time"
	
	"github.com/maxnordlund/breamio/briee"
)

func mockStandard(t float64) (float64, float64) {
	return math.Cos(t), math.Sin(t)
}

func mockConstant(t float64) (float64, float64) {
	return 0.5, 0.5
}

type MockDriver struct{}

func (d MockDriver) List() []string {
	return []string{"standard", "constant"}
}

func (d MockDriver) Create() (Tracker, error) {
	return &MockTracker{mockStandard, 0, false, false}, nil
}
func (d MockDriver) CreateFromId(identifier string) (Tracker, error) {
	switch identifier {
	case "standard":
		return d.Create()
	case "constant":
		return &MockTracker{mockConstant, 0, false, false}, nil
	default:
		return nil, errors.New("No such tracker.")
	}
}

type MockTracker struct {
	f                     func(float64) (float64, float64)
	t                     float64
	connected, calibrated bool
}

func (m *MockTracker) Stream() (<-chan *ETData, <-chan error) {
	ch := make(chan *ETData)
	errs := make(chan error, 1)
	go m.generate(ch)
	return ch, errs
}

func (m *MockTracker) Link(ee briee.EventEmitter) {
	ch := ee.Publish("tracker:etdata", &ETData{}).(chan<- *ETData)
	go m.generate(ch)
}

func (m *MockTracker) Close() error {
	m.connected = false
	return nil
}

func (m *MockTracker) Connect() error {
	m.t = 0
	m.connected = true
	return nil
}

func (m MockTracker) Calibrate(points <-chan Point2D, errs chan<- error) {
	errs <- NotImplementedError("Calibrate of MockTracker")
}

func (m MockTracker) IsCalibrated() bool {
	return m.calibrated
}

func (m *MockTracker) generate(ch chan<- *ETData) {
	for m.connected {
		x, y := m.f(m.t)
		ch <- &ETData{point2D{x, y}, time.Now()}
		m.t += 0.1
	}
}

func init() {
	drivers["mock"] = new(MockDriver)
}
