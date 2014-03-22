package mock

import (
	"errors"
	"math"
	"time"

	"github.com/maxnordlund/breamio/briee"
	. "github.com/maxnordlund/breamio/gorgonzola"
)

func mockStandard(t float64) (float64, float64) {
	return 0.5 + 0.5*math.Cos(t), 0.5 + 0.5*math.Sin(t)
}

func mockConstant(t float64) (float64, float64) {
	return 0.5, 0.5
}

type MockDriver struct{}

func (d MockDriver) List() []string {
	return []string{"standard", "constant"}
}

func (d MockDriver) Create() (Tracker, error) {
	return New(mockStandard), nil
}
func (d MockDriver) CreateFromId(identifier string) (Tracker, error) {
	switch identifier {
	case "standard":
		return New(mockStandard), nil
	case "constant":
		return New(mockConstant), nil
	default:
		return nil, errors.New("No such tracker.")
	}
}

type MockTracker struct {
	f                     func(float64) (float64, float64)
	t                     float64
	calibrated            bool
	calibrationPoints     int
	validationPoints      int
	closer                chan struct{}
}

func New(f func(float64) (float64, float64)) *MockTracker {
	return &MockTracker{f, 0, false, 0, 0, make(chan stuct{})}
}

func (m *MockTracker) Stream() (<-chan *ETData, <-chan error) {
	ch := make(chan *ETData)
	errs := make(chan error, 1)
	go m.generate(ch)
	return ch, errs
}

func (m *MockTracker) Link(ee briee.EventEmitter) {
	etDataCh := ee.Publish("tracker:etdata", &ETData{}).(chan<- *ETData)
	go m.generate(etDataCh)
	m.setupCalibrationEvents(ee)
}

func (m *MockTracker) setupCalibrationEvents(ee briee.EventEmitter) {
	go m.calibrateStartHandler(ee)
	go m.calibrateAddHandler(ee)
	
	go m.validateStartHandler(ee)
	go m.validateAddHandler(ee)
}

func (m *MockTracker) Close() error {
	m.connected = false
	close(m.closer)
	return nil
}

func (m *MockTracker) Connect() error {
	m.t = 0
	m.connected = true
	m.closer = make(chan struct{})
	return nil
}

func (m MockTracker) IsCalibrated() bool {
	return m.calibrated
}

func (m *MockTracker) generate(ch chan<- *ETData) {
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for t := range ticker.C {
		x, y := m.f(m.t)
		select {
			case ch <- &ETData{Point2D{x, y}, t}:
			case <-m.closer: return
			default:
		}
		m.t += 0.01
	}
}

func (m *MockTracker) calibrateStartHandler(ee briee.EventEmitter) {
	inCh := ee.Subscribe("tracker:calibrate:start", struct{}{}).(<-chan struct{})
	outCh := ee.Publish("tracker:calibrate:next", struct{}{}).(chan<- struct{})
	defer ee.Unsubscribe("tracker:calibrate:next", outCh)
	
	for {
		select {
			case <- inCh:
				m.calibrating = true
				m.calibrationPoints = 0
				outCh <- struct{}{}
			case <-m.closer: return
			default:
		}
	}
}

func (m *MockTracker) calibrateAddHandler(ee briee.EventEmitter) {
	inCh := ee.Subscribe("tracker:calibrate:add", Point2D{}).(<-chan struct{})
	
	nextCh := ee.Publish("tracker:calibrate:next", struct{}{}).(chan<- struct{})
	defer ee.Unsubscribe("tracker:calibrate:next", outCh)
	
	endCh := ee.Publish("tracker:calibrate:end", struct{}{}).(chan<- struct{})
	defer ee.Unsubscribe("tracker:calibrate:next", outCh)
	
	for {
		select {
			case <- inCh:
				m.calibrationPoints++

				if m.calibrationPoints >= 5 {
					endCh <- struct{}{}
					ee.Dispatch("tracker:validate:start", struct{}{})
				} else {
					nextCh <- struct{}{}
				}
			case <-m.closer: return
			default:
		}
	}
}

func (m *MockTracker) validateStartHandler(ee briee.EventEmitter) {
	inCh := ee.Subscribe("tracker:validate:start", struct{}{}).(<-chan struct{})
	outCh := ee.Publish("tracker:validate:next", struct{}{}).(chan<- struct{})
	defer ee.Unsubscribe("tracker:validate:next", outCh)
	
	for {
		select {
			case <- inCh:
				m.calibrating = true
				m.validationPoints = 0
				outCh <- struct{}{}
			case <-m.closer: return
			default:
		}
	}
}

func (m *MockTracker) validateAddHandler(ee briee.EventEmitter) {
	inCh := ee.Subscribe("tracker:validate:add", Point2D{}).(<-chan struct{})
	
	nextCh := ee.Publish("tracker:validate:next", struct{}{}).(chan<- struct{})
	defer ee.Unsubscribe("tracker:validate:next", outCh)
	
	qualityCh := ee.Publish("tracker:validate:quality", float64(0)).(chan<- struct{})
	defer ee.Unsubscribe("tracker:validate:next", outCh)
	
	for {
		select {
			case <- inCh:
				m.validationPoints++
				
				if m.validationPoints >= 5 {
					endCh <- struct{}{}
					qualityCh <- float64(0.05)
				} else {
					nextCh <- struct{}{}
				}
			case <-m.closer: return
			default:
		}
	}
}

func init() {
	RegisterDriver("mock", new(MockDriver))
}
