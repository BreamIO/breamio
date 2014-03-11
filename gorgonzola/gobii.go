package gorgonzola

import (
	"github.com/zephyyrr/gobii/gaze"
	"github.com/maxnordlund/breamio/briee"
)

// Driver implementation for Gobii
type GazeDriver struct{}

func (GazeDriver) Create() (Tracker, error) {
	tracker, err := gaze.AnyEyeTracker()
	return &GazeTracker{tracker, false}, err
}

func (GazeDriver) CreateFromId(id string) (Tracker, error) {
	//TODO fix this to use the real function
	tracker, err := gaze.EyeTrackerFromURL("tet-usb://" + id)
	return &GazeTracker{tracker, false}, err
}

func (GazeDriver) List() (res []string) {
	list, err := gaze.USBTrackers()
	res = make([]string, 0, len(list))
	if err != nil {
		return
	}
	for _, info := range list {
		res = append(res, info.SerialNumber())
	}
	return
}

type GazeTracker struct {
	*gaze.EyeTracker
	calibrated bool
}

func (g GazeTracker) Stream() (<-chan *ETData, <-chan error) {
	ch := make(chan *ETData)
	errs := make(chan error, 1)

	err := g.StartTracking(gobiiOnGazeCallback(ch))

	if err != nil {
		errs <- err
	}
	return ch, errs
}

func (g *GazeTracker) Link(ee briee.EventEmitter) {
	etdataCh := ee.Publish("tracker:etdata", &ETData{}).(chan<- *ETData)
	err := g.StartTracking(gobiiOnGazeCallback(etdataCh));
	ee.Dispatch("tracker:error", err);
}

func (g *GazeTracker) Calibrate(points <-chan Point2D, errors chan<- error) {
	for _ = range points {
		//Mock implementation until Calibration implemented in gobii
		errors <- NotImplementedError("Calibrate")
	}
	g.calibrated = true
}

func (g GazeTracker) IsCalibrated() bool {
	return g.calibrated
}

func gobiiOnGazeCallback(ch chan<-*ETData) func(data *gaze.GazeData) {
	return func(data *gaze.GazeData) {
		etdata := new(ETData)
		etdata.Filtered = filter(data.Left().GazePointOnDisplay(), data.Right().GazePointOnDisplay())
		etdata.Timestamp = data.Timestamp()
		ch <- etdata
	}
}

func init() {
	drivers["gobii"] = new(GazeDriver)
}
