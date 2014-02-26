package gorgonzola

type Gobii struct{
	gobii.EyeTracker
	calibrated bool
}

func (g Gobii) Stream() (ch <-chan *ETData, errs <-chan error) {
	ch = make(chan *ETData)
	errs = make(chan error, 1)

	err := g.StartTracking(func(data *GazeData) {
		etdata := new(ETData)
		etdata.filtered.X = (data.Left().GazeDataOnDisplay().X() + data.Right().GazeDataOnDisplay().X())/2
		etdata.filtered.Y = (data.Left().GazeDataOnDisplay().Y() + data.Right().GazeDataOnDisplay().Y())/2
		etdata.timestamp0 = data.Timestamp()
		ch <- etdata
	})

	if err != nil {
		errs <- err
	}
	return
}
	
func (g *Gobii) Calibrate(points <-chan Point2D, errors chan<- error) {
	for p := range points{
		//Mock implementation until Calibration implemented in gobii
		errors<- NotImplemented("Calibrate")
	}
	g.calibrated = true
}

func (g Gobii) IsCalibrated() bool {
	return g.calibrated
}