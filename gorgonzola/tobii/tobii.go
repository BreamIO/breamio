package tobii

import (
	"fmt"
	//"log"
	"github.com/maxnordlund/breamio/briee"
	. "github.com/maxnordlund/breamio/gorgonzola"
	"github.com/zephyyrr/gobii/gaze"
)

// Driver implementation for Gobii
type GazeDriver struct{}

func (GazeDriver) Create() (Tracker, error) {
	tracker, err := gaze.AnyEyeTracker()
	return &GazeTracker{tracker, false}, err
}

func (g GazeDriver) CreateFromId(id string) (Tracker, error) {
	url := "tet-usb://" + id
	tracker, err := gaze.EyeTrackerFromURL(url)
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

func (g *GazeTracker) Link(ee briee.PublishSubscriber) {
	etdataCh := ee.Publish("tracker:etdata", &ETData{}).(chan<- *ETData)
	err := g.StartTracking(gobiiOnGazeCallback(etdataCh))
	errorCh := ee.Publish("tracker:error", err.(error)).(chan<- error)
	//defer close(errorCh)
	errorCh <- err
}

func (g *GazeTracker) Calibrate(points <-chan XYer, errors chan<- error) {
	for _ = range points {
		//Mock implementation until Calibration implemented in gobii
		errors <- NotImplementedError("Calibrate")
	}
	g.calibrated = true
}

func (g GazeTracker) IsCalibrated() bool {
	return g.calibrated
}

func (g GazeTracker) String() string {
	return fmt.Sprintf("<GobiiTracker %v>", g.EyeTracker)
}

func gobiiOnGazeCallback(ch chan<- *ETData) func(data *gaze.GazeData) {
	return func(data *gaze.GazeData) {
		ts := data.TrackingStatus()
		if ts < gaze.BothEyesTracked || ts == gaze.OneEyeTrackedUnknownWhich {
			return //Bad data
		}
		etdata := new(ETData)
		etdata.Filtered = Filter(data.Left().GazePointOnDisplay(), data.Right().GazePointOnDisplay())
		etdata.Timestamp = data.Timestamp()
		//log.Println(etdata)
		ch <- etdata
	}
}

func init() {
	RegisterDriver("tobii", new(GazeDriver))
}
