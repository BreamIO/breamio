package eyetribe

import (
	"github.com/maxnordlund/breamio/briee"
	"github.com/maxnordlund/breamio/gorgonzola"
	"github.com/zephyyrr/thegotribe"
	"io"
)

func init() {
	gorgonzola.RegisterDriver("tribe", driver{})
}

type driver struct {
}

func (d driver) Create() (gorgonzola.Tracker, error) {
	return d.CreateFromId(thegotribe.DefaultAddr)
}

func (d driver) CreateFromId(id string) (gorgonzola.Tracker, error) {
	return nil, nil
}

func (d driver) List() []string {
	return []string{thegotribe.DefaultAddr}
}

type tracker struct {
	*thegotribe.EyeTracker
}

// Returns a channel of tracking data.
// A error channel is also returned.
// As the tracker reads data from the device,
// new ETDatas is created and sent on the channel.
// If the channel is full, the tracker discards the data.
// If a error occurs while streaming, it is sent along the error channel.
func (t tracker) Stream() (<-chan *gorgonzola.ETData, <-chan error) {
	data, errs := make(chan *gorgonzola.ETData), make(chan error)
	go func() {
		defer close(data)
		for f := range t.EyeTracker.Frames {
			data <- &gorgonzola.ETData{
				Filtered:  *gorgonzola.ToPoint2D(p2d{f.Average}),
				LeftGaze:  gorgonzola.ToPoint2D(p2d{f.LeftEye.Raw}),
				RightGaze: gorgonzola.ToPoint2D(p2d{f.RightEye.Raw}),
				Timestamp: f.Timestamp.Time,
			}
		}
		errs <- io.EOF
	}()
	return data, errs
}

// Connects the tracker to the given Event Emitter
// This means the tracker publishes its capabilities to the emitter.
// That means that gaze data is published to the channel on event "tracker:etdata".
// It also means calibration is exposed on "tracker:calibrate" and any settings capabilities on tracker:settings
func (t tracker) Link(ee briee.PublishSubscriber) {
	data := ee.Publish("tracker:etdata", &gorgonzola.ETData{}).(chan<- *gorgonzola.ETData)
	shutdown := ee.Subscribe("shutdown", struct{}{}).(<-chan struct{})
	tshutdown := ee.Subscribe("tracker:shutdown", struct{}{}).(<-chan struct{})

	stream, _ := t.Stream()
	// The only error expected to show up is io.EOF.
	// This also closes the stream, and everything comes to a natural halt.
	go func() {
		defer close(data)
		defer ee.Unsubscribe("tracker:shutdown", tshutdown)
		defer ee.Unsubscribe("shutdown", shutdown)
		for {
			select {
			case d, ok := <-stream:
				if ok {
					data <- d
				} else {
					return
				}
			case <-tshutdown:
				return
			case <-shutdown:
				return
			}
		}

	}()
}

//Initiates a connection between the software driver and the hardware.
//Should be called before any other use of the tracker except if method specifies it.
func (t tracker) Connect() error {
	var err error
	t.EyeTracker, err = thegotribe.Create()
	return err
}

//Closes the tracker connection and performs any other clean up necessary in the driver.
func (t tracker) Close() error {
	return t.EyeTracker.Close()
}
