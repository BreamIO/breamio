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
	return &GazeTracker{
		tracker,
		make(chan struct{}),
		nil,
		false,
		0,
		0,
	}, err
}

func (g GazeDriver) CreateFromId(id string) (Tracker, error) {
	url := "tet-usb://" + id
	tracker, err := gaze.EyeTrackerFromURL(url)
	return &GazeTracker{tracker, make(chan struct{}), nil, false, 0, 0}, err
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
	gaze.EyeTracker
	closer            chan struct{}
	etdataCh          chan<- *ETData
	calibrated        bool
	calibrationPoints uint
	validationPoints  uint
}

func (g GazeTracker) Stream() (<-chan *ETData, <-chan error) {
	ch := make(chan *ETData)
	g.etdataCh = (chan<- *ETData)(ch)
	errs := make(chan error, 1)

	err := g.StartTracking(gobiiOnGazeCallback(g.etdataCh))

	if err != nil {
		errs <- err
	}
	return ch, errs
}

func (g *GazeTracker) Link(ee briee.PublishSubscriber) {
	g.etdataCh = ee.Publish("tracker:etdata", &ETData{}).(chan<- *ETData)
	defer RemoveTracker(g)

	go g.setupCalibrationEvents(ee)
	go func() {
		shutdownCh := ee.Subscribe("shutdown", struct{}{}).(<-chan struct{})
		tShutdownCh := ee.Subscribe("tracker:shutdown", struct{}{}).(<-chan struct{})
		defer ee.Unsubscribe("shutdown", shutdownCh)
		defer ee.Unsubscribe("tracker:shutdown", tShutdownCh)
		select {
		case <-shutdownCh:
		case <-tShutdownCh:
		}
		close(g.closer)
	}()

	err := g.StartTracking(gobiiOnGazeCallback(g.etdataCh))
	if err != nil {
		errorCh := ee.Publish("tracker:error", NewError("")).(chan<- Error)
		//defer close(errorCh)
		errorCh <- NewError(err.Error())
	}
}

func (g *GazeTracker) Close() error {
	close(g.closer)
	defer close(g.etdataCh) //We want to do this after closing the tracker, but we want to return the error from closing. This saves on temporary variables.
	return g.EyeTracker.Close()
}

func (g *GazeTracker) setupCalibrationEvents(ee briee.PublishSubscriber) {
	go g.calibrateStartHandler(ee)
	go g.calibrateAddHandler(ee)

	go g.validateStartHandler(ee)
	go g.validateAddHandler(ee)
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
		etdata.Timestamp = data.Timestamp()
		etdata.Filtered = *ToPoint2D(Filter(data.Left().GazePointOnDisplay(), data.Right().GazePointOnDisplay()))
		etdata.LeftGaze = ToPoint2D(data.Left().GazePointOnDisplay())
		etdata.RightGaze = ToPoint2D(data.Right().GazePointOnDisplay())
		//log.Println(etdata)
		ch <- etdata
	}
}

func handleError(errorCh chan<- Error, f func()) func(error) {
	return func(err error) {
		if err != nil {
			errorCh <- NewError(err.Error())
			return
		}
		f()
	}
}

func (g *GazeTracker) calibrateStartHandler(ee briee.PublishSubscriber) {
	inCh := ee.Subscribe("tracker:calibrate:start", struct{}{}).(<-chan struct{})
	outCh := ee.Publish("tracker:calibrate:next", struct{}{}).(chan<- struct{})
	errorCh := ee.Publish("tracker:calibrate:error", NewError("")).(chan<- Error)
	defer ee.Unsubscribe("tracker:calibrate:start", outCh)
	defer close(outCh)

	for {
		select {
		case <-inCh:
			g.StartCalibration(handleError(errorCh, func() {
				g.calibrationPoints = 0
				outCh <- struct{}{}
			}))
		case <-g.closer:
			return
		}
	}
}

func (g *GazeTracker) calibrateAddHandler(ee briee.PublishSubscriber) {
	inCh := ee.Subscribe("tracker:calibrate:add", Point2D{}).(<-chan Point2D)
	defer ee.Unsubscribe("tracker:calibrate:add", inCh)

	nextCh := ee.Publish("tracker:calibrate:next", struct{}{}).(chan<- struct{})
	defer close(nextCh)

	endCh := ee.Publish("tracker:calibrate:end", struct{}{}).(chan<- struct{})
	defer close(endCh)

	vstartCh := ee.Publish("tracker:validate:start", struct{}{}).(chan<- struct{})
	defer close(vstartCh)

	errorCh := ee.Publish("tracker:calibrate:error", NewError("")).(chan<- Error)
	defer close(errorCh)

	for {
		select {
		case p := <-inCh:
			g.calibrationPoints++
			//println("calibration points:", g.calibrationPoints)
			if g.calibrationPoints >= 5 {
				computed := make(chan struct{})
				g.ComputeAndSetCalibration(handleError(errorCh, func() {
					close(computed)
				}))

				<-computed

				g.StopCalibration(handleError(errorCh, func() {
					endCh <- struct{}{}
					vstartCh <- struct{}{}
				}))

			} else {
				g.AddPointToCalibration(gaze.NewPoint2D(p.X(), p.Y()),
					handleError(errorCh, func() {
						nextCh <- struct{}{}
					}))

			}
		case <-g.closer:
			return
		}
	}
}

func (g *GazeTracker) validateStartHandler(ee briee.PublishSubscriber) {
	inCh := ee.Subscribe("tracker:validate:start", struct{}{}).(<-chan struct{})
	nextCh := ee.Publish("tracker:validate:next", struct{}{}).(chan<- struct{})
	defer ee.Unsubscribe("tracker:validate:start", inCh)
	defer close(nextCh)

	for {
		select {
		case <-inCh:
			g.validationPoints = 0
			nextCh <- struct{}{}
		case <-g.closer:
			return
		}
	}
}

//TODO do actual implementation.
func (g *GazeTracker) validateAddHandler(ee briee.PublishSubscriber) {
	inCh := ee.Subscribe("tracker:validate:add", Point2D{}).(<-chan Point2D)
	defer ee.Unsubscribe("tracker:validate:add", inCh)

	nextCh := ee.Publish("tracker:validate:next", struct{}{}).(chan<- struct{})
	defer close(nextCh)

	qualityCh := ee.Publish("tracker:validate:end", float64(0)).(chan<- float64)
	defer close(qualityCh)

	for {
		select {
		case <-inCh:
			g.validationPoints++
			if g.validationPoints >= 5 {
				//Calculate this using the tobiigaze_get_calibration_point_data_items instead.
				qualityCh <- float64(0.05)
			} else {
				nextCh <- struct{}{}
			}
		case <-g.closer:
			return
		}
	}
}

func init() {
	RegisterDriver("tobii", new(GazeDriver))
}
