package gorgonzola

import (
	"github.com/zephyyrr/gobii/gaze"
)

// Driver implementation for Gobii
type GazeDriver struct{}

func (GazeDriver) Create() (Tracker, error) {
	tracker,err := gaze.AnyEyeTracker()
	return &GazeTracker{tracker, false}, err
}

func (GazeDriver) CreateS(id string) (Tracker, error) {
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

type GazeTracker struct{
	*gaze.EyeTracker
	calibrated bool
}

func (g GazeTracker) Stream() (<-chan *ETData,<-chan error) {
	ch := make(chan *ETData)
	errs := make(chan error, 1)

	err := g.StartTracking(func(data *gaze.GazeData) {
		etdata := new(ETData)
		etdata.filtered.X = (data.Left().GazePointOnDisplay().X() + data.Right().GazePointOnDisplay().X())/2
		etdata.filtered.Y = (data.Left().GazePointOnDisplay().Y() + data.Right().GazePointOnDisplay().Y())/2
		etdata.timestamp = data.Timestamp()
		ch <- etdata
	})

	if err != nil {
		errs <- err
	}
	return ch, errs
}
	
func (g *GazeTracker) Calibrate(points <-chan Point2D, errors chan<- error) {
	for _ = range points{
		//Mock implementation until Calibration implemented in gobii
		errors<- NotImplementedError("Calibrate")
	}
	g.calibrated = true
}

func (g GazeTracker) IsCalibrated() bool {
	return g.calibrated
}

func init() {
	drivers["gobii"] = new(GazeDriver)
}