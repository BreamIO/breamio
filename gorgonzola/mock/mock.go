package mock

import (
	"errors"
	"log"
	"math"
	"math/rand"
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

func mockSporadic(t float64) (float64, float64) {
	return 0.5 + 0.5*math.Cos(t) + rand.Float64()/50, 0.5 + 0.5*math.Sin(10*t) + rand.Float64()/50
}

var prevX, prevY float64
func mockRandomFixation(t float64) (float64, float64) {
	var retX, retY float64

	// Stay or go?
	if math.Abs(rand.NormFloat64()) <= 2.0 {
		// Stay
		retX = prevX + rand.NormFloat64() * 0.01
		retY = prevY + rand.NormFloat64() * 0.01
	} else {
		// Go
		//dx = math.Cos(2*3.1415*rand.Float64())/5.0
		//dy = math.Sin(2*3.1415*rand.Float64())/5.0
		retX = rand.Float64()
		retY = rand.Float64()
	}

	prevX = retX
	prevY = retY

	return retX, retY
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
	case "sporadic":
		return New(mockSporadic), nil
	case "random_fix":
		return New(mockRandomFixation), nil
	default:
		return nil, errors.New("No such tracker.")
	}
}

type MockTracker struct {
	f                 func(float64) (float64, float64)
	t                 float64
	calibrating       bool
	calibrated        bool
	calibrationPoints int
	validationPoints  int
	closer            chan struct{}
}

func New(f func(float64) (float64, float64)) *MockTracker {
	return &MockTracker{f, 0, false, false, 0, 0, nil}
}

func (m *MockTracker) Stream() (<-chan *ETData, <-chan error) {
	ch := make(chan *ETData)
	errs := make(chan error, 1)
	go m.generate(ch)
	return ch, errs
}

func (m *MockTracker) Link(ee briee.PublishSubscriber) {
	etDataCh := ee.Publish("tracker:etdata", &ETData{}).(chan<- *ETData)
	go m.generate(etDataCh)
	m.setupCalibrationEvents(ee)

	go func() {
		defer RemoveTracker(m)
		shutdownCh := ee.Subscribe("shutdown", struct{}{}).(<-chan struct{})
		tShutdownCh := ee.Subscribe("tracker:shutdown", struct{}{}).(<-chan struct{})
		defer ee.Unsubscribe("shutdown", shutdownCh)
		defer ee.Unsubscribe("tracker:shutdown", tShutdownCh)
		select {
		case <-shutdownCh:
		case <-tShutdownCh:
		}
		m.Close()
	}()

}

func (m *MockTracker) setupCalibrationEvents(ee briee.PublishSubscriber) {
	go m.calibrateStartHandler(ee)
	go m.calibrateAddHandler(ee)

	go m.validateStartHandler(ee)
	go m.validateAddHandler(ee)
}

func (m *MockTracker) Close() error {
	close(m.closer)
	return nil
}

func (m *MockTracker) Connect() error {
	m.t = 0
	m.closer = make(chan struct{})
	return nil
}

func (m *MockTracker) generate(ch chan<- *ETData) {
	ticker := time.NewTicker(REFRESH * time.Millisecond)
	defer ticker.Stop()
	for t := range ticker.C {
		select {
		case <-m.closer:
			close(ch)
			return
		default:
		}
		x, y := m.f(m.t)
		ch <- &ETData{
			Filtered:  Point2D{x, y},
			Timestamp: t,
		}
		m.t += 0.01
	}
}

func (m *MockTracker) calibrateStartHandler(ee briee.PublishSubscriber) {
	inCh := ee.Subscribe("tracker:calibrate:start", struct{}{}).(<-chan struct{})
	outCh := ee.Publish("tracker:calibrate:next", struct{}{}).(chan<- struct{})
	defer ee.Unsubscribe("tracker:calibrate:start", outCh)
	defer close(outCh)
	for {
		select {
		case <-inCh:
			log.Println("MockTracker#calibrateStartHandler", "tracker:calibrate:start")
			m.calibrating = true
			m.calibrationPoints = 0
			outCh <- struct{}{}
		case <-m.closer:
			return
		}
	}
}

func (m *MockTracker) calibrateAddHandler(ee briee.PublishSubscriber) {
	inCh := ee.Subscribe("tracker:calibrate:add", Point2D{}).(<-chan Point2D)
	defer ee.Unsubscribe("tracker:calibrate:add", inCh)

	nextCh := ee.Publish("tracker:calibrate:next", struct{}{}).(chan<- struct{})
	defer close(nextCh)

	endCh := ee.Publish("tracker:calibrate:end", struct{}{}).(chan<- struct{})
	defer close(endCh)

	vstartCh := ee.Publish("tracker:validate:start", struct{}{}).(chan<- struct{})
	defer close(vstartCh)

	for {
		select {
		case <-inCh:
			log.Println("MockTracker#calibrateAddHandler", "tracker:calibrate:add")
			m.calibrationPoints++
			if m.calibrationPoints >= 5 {
				endCh <- struct{}{}
				vstartCh <- struct{}{}
			} else {
				nextCh <- struct{}{}
			}
		case <-m.closer:
			return
		}
	}
}

func (m *MockTracker) validateStartHandler(ee briee.PublishSubscriber) {
	inCh := ee.Subscribe("tracker:validate:start", struct{}{}).(<-chan struct{})
	nextCh := ee.Publish("tracker:validate:next", struct{}{}).(chan<- struct{})
	defer ee.Unsubscribe("tracker:validate:start", inCh)
	//defer close(nextCh)

	for {
		select {
		case <-inCh:
			log.Println("MockTracker#validateStartHandler", "tracker:validate:start")
			m.calibrating = true
			m.validationPoints = 0
			nextCh <- struct{}{}
		case <-m.closer:
			return
		}
	}
}

func (m *MockTracker) validateAddHandler(ee briee.PublishSubscriber) {
	inCh := ee.Subscribe("tracker:validate:add", Point2D{}).(<-chan Point2D)
	defer ee.Unsubscribe("tracker:validate:add", inCh)

	nextCh := ee.Publish("tracker:validate:next", struct{}{}).(chan<- struct{})
	//defer close(nextCh)

	qualityCh := ee.Publish("tracker:validate:end", float64(0)).(chan<- float64)
	//defer close(qualityCh)

	for {
		select {
		case <-inCh:
			log.Println("MockTracker#validateAddHandler", "tracker:validate:add")
			m.validationPoints++
			if m.validationPoints >= 5 {
				qualityCh <- float64(0.05)
			} else {
				nextCh <- struct{}{}
			}
		case <-m.closer:
			return
		}
	}
}

func init() {
	RegisterDriver("mock", new(MockDriver))
}
