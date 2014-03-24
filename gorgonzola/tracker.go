package gorgonzola

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/maxnordlund/breamio/briee"
)

var drivers = make(map[string]Driver)

func GetDriver(typ string) Driver {
	return drivers[typ]
}

func RegisterDriver(typ string, driver Driver) error {
	if drivers[typ] != nil {
		return errors.New(fmt.Sprintf("%s is already registered", typ))
	}
	if driver == nil {
		return errors.New("Nil implementations is not allowed.")
	}
	drivers[typ] = driver
	return nil
}

//Drivers specify tracker operations not connected to any specific tracker and constructors
//
//In order to allow multiple different implementations of trackers (Tobii, Mirametrix, Mock, TheEyeTrive) simultaneously, we can not specify all of these in all functions.
//Instead of hard coding all possible, we can allow injection of new types that follow a common interface.
//These can then be iterated and processed.
type Driver interface {
	//Creates any tracker from this driver.
	//No promises are made to the uniqueness of the tracker returned.
	//If no tracker can be returned, a error is returned instead.
	Create() (Tracker, error)

	//Creates a tracker connected to the identifier string.
	//The driver is obliged to return that tracker and only that tracker.
	//If the identifier is invalid or no longer connected, a error is returned.
	CreateFromId(identifier string) (Tracker, error)

	//Returns a list of valid identifiers that can be used with CreateS.
	//Empty if no trackers can be created.
	List() []string
}

//A common interface for all trackers.
type Tracker interface {
	// Returns a channel of tracking data.
	// A error channel is also returned.
	// As the tracker reads data from the device,
	// new ETDatas is created and sent on the channel.
	// If the channel is full, the tracker discards the data.
	// If a error occurs while streaming, it is sent along the error channel.
	Stream() (<-chan *ETData, <-chan error)

	// Connects the tracker to the given Event Emitter
	// This means the tracker publishes its capabilities to the emitter.
	// That means that gaze data is published to the channel on event "tracker:etdata".
	// It also means calibration is exposed on "tracker:calibrate" and any settings capabilities on tracker:settings
	Link(ee briee.PublishSubscriber)

	//Initiates a connection between the software driver and the hardware.
	//Should be called before any other use of the tracker except if method specifies it.
	Connect() error

	//Closes the tracker connection and performs any other clean up necessary in the driver.
	io.Closer
}

type ETData struct {
	Filtered  XYer
	Timestamp time.Time
}

//Lists all trackers reported by all trackers
func List() []string {
	res := make([]string, 0, 32)
	for _, driver := range drivers {
		res = append(res, driver.List()...)
	}
	return res
}

//Error type used for signaling a unimplemented method.
type NotImplementedError string

func (e NotImplementedError) Error() string {
	return fmt.Sprintf("%s is not implemented.", string(e))
}
