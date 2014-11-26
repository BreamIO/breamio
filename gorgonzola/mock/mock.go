package mock

import (
	"errors"
	"math"
	"math/rand"
	"time"

	"github.com/maxnordlund/breamio/beenleigh"
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
		retX = prevX + rand.NormFloat64()*0.01
		retY = prevY + rand.NormFloat64()*0.01
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

func mockConstantFixation(t float64) (float64, float64) {
	return 0.5 + rand.NormFloat64()*0.01, 0.5 + rand.NormFloat64()*0.01
}

var log beenleigh.Logger = beenleigh.NewLoggerS("MockDriver")

type MockDriver struct{}

func (d MockDriver) List() []string {
	return []string{"standard", "constant"}
}

func (d MockDriver) Create(c beenleigh.Constructor) (Tracker, error) {
	return New(c, mockStandard), nil
}
func (d MockDriver) CreateFromId(c beenleigh.Constructor, identifier string) (Tracker, error) {
	switch identifier {
	case "standard":
		return New(c, mockStandard), nil
	case "constant":
		return New(c, mockConstant), nil
	case "sporadic":
		return New(c, mockSporadic), nil
	case "fixations":
		return New(c, mockRandomFixation), nil
	case "constfix":
		return New(c, mockConstantFixation), nil
	default:
		return nil, errors.New("No such tracker.")
	}
}

type MockTracker struct {
	beenleigh.SimpleModule
	f                 func(float64) (float64, float64)
	t                 float64
	calibrating       bool
	calibrated        bool
	calibrationPoints int
	validationPoints  int
	closer            chan struct{}

	MethodCalibrateStart beenleigh.EventMethod `event:"tracker:calibrate:start" returns:"calibrate:next"`
	MethodCalibrateAdd   beenleigh.EventMethod `event:"tracker:calibrate:add" returns:"calibrate:next,calibrate:end,validate:start"`
	MethodValidateStart  beenleigh.EventMethod `event:"tracker:validate:start" returns:"validate:next"`
	MethodValidateAdd    beenleigh.EventMethod `event:"tracker:validate:add" returns:"validate:next,validate:end"`
}

func New(c beenleigh.Constructor, f func(float64) (float64, float64)) *MockTracker {
	mt := &MockTracker{
		SimpleModule: beenleigh.NewSimpleModule("tracker", c),
		f:            f,
		closer:       make(chan struct{}),
	}

	emitter := c.Logic.CreateEmitter(c.Emitter)
	mt.Link(emitter)
	return mt
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
}

func (m *MockTracker) Close() error {
	close(m.closer)
	return nil
}

func (m *MockTracker) Connect() error {
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

func (m *MockTracker) CalibrateStart() beenleigh.Signal {
	log.Println("MockTracker#calibrateStartHandler", "tracker:calibrate:start")
	m.calibrating = true
	m.calibrationPoints = 0
	return beenleigh.Pulse
}

func (m *MockTracker) CalibrateAdd(p Point2D) (next, end, validate beenleigh.Signal) {
	log.Println("MockTracker#calibrateAddHandler", "tracker:calibrate:add")
	m.calibrationPoints++
	if m.calibrationPoints >= 5 {
		return nil, beenleigh.Pulse, beenleigh.Pulse
	} else {
		return beenleigh.Pulse, nil, nil
	}
}

func (m *MockTracker) ValidateStart() beenleigh.Signal {
	log.Println("MockTracker#validateStartHandler", "tracker:validate:start")
	m.calibrating = true
	m.validationPoints = 0
	return beenleigh.Pulse
}

func (m *MockTracker) ValidateAdd(p Point2D) (next beenleigh.Signal, end *float64) {
	log.Println("MockTracker#validateAddHandler", "tracker:validate:add")
	m.validationPoints++
	if m.validationPoints >= 5 {
		ans := 0.05
		return nil, &ans
	} else {

		return beenleigh.Pulse, nil
	}
}

func init() {
	RegisterDriver("mock", new(MockDriver))
}
