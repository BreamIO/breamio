package gorgonzola

import (
	"errors"
	bl "github.com/maxnordlund/breamio/beenleigh"
	"log"
	"os"
	"strings"
)

type GorgonzolaRun struct {
	closing chan struct{}
}

func (gr GorgonzolaRun) Run(logic bl.Logic) {
	newCh := logic.RootEmitter().Subscribe("new:tracker", bl.Spec{}).(<-chan bl.Spec)
	for {
		select {
		case <-gr.closing:
			return
		case spec := <-newCh:
			if err := gr.onNewEvent(logic, spec); err != nil {
				logic.RootEmitter().Dispatch("gorgonzola:error", err)
			}
		}
	}
}

func (gr GorgonzolaRun) Close() error {
	close(gr.closing)
	return nil
}

func (gr GorgonzolaRun) onNewEvent(logic bl.Logic, event bl.Spec) error {
	logger.Println("Recieved new:tracker event.")

	tracker, err := CreateFromURI(event.Data)
	if err != nil {
		logger.Printf("Could not create new tracker with uri %s: %s", event.Data, err)
		return err
	}
	err = tracker.Connect()
	if err != nil {
		logger.Println("Unable to connect to tracker:", err)
		return err
	}

	ee := logic.CreateEmitter(event.Emitter)
	go tracker.Link(ee)

	logger.Printf("Created a new tracker with uri %s on EE %d.\n", event.Data, event.Emitter)
	return nil
}

func init() {
	bl.Register(&GorgonzolaRun{make(chan struct{})})
}

var logger = log.New(os.Stdout, "[Gorgonzola]", log.LstdFlags)

// Creates a tracker using a URI.
// The URI is on the form <driver>://<id>.
// The driver part is used to find a registered driver in the tracker driver table.
// The id part is used as a argument for the selected drivers CreateFromId method.
func CreateFromURI(uri string) (Tracker, error) {
	split := strings.SplitN(uri, "://", 2)
	if len(split) < 2 {
		return nil, errors.New("Malformed URI.")
	}
	typ, id := split[0], split[1]
	driver := GetDriver(typ)
	if driver == nil {
		return nil, errors.New("No such driver.")
	}
	return GetDriver(typ).CreateFromId(id)
}

// A interface specifying something that can deliver
// x and y coordinates in Cartesian space.
type XYer interface {
	X() float64
	Y() float64
}

// A struct of two float64s.
// It is meant to serve as a practical implementation of XYer.
type Point2D struct {
	Xf, Yf float64
}

// Returns the X part of the Point2D
func (p Point2D) X() float64 {
	return p.Xf
}

// Returns the Y part of the Point2D
func (p Point2D) Y() float64 {
	return p.Yf
}

func Filter(left, right XYer) XYer {
	return Point2D{(left.X() + right.X()) / 2, (left.Y() + right.Y()) / 2}
}

// Due to the lack of interface types in the briee.EventEmitter events,
// Gorgonzola needs a common error type.
// This is that type.
type Error struct {
	err string
}

// Constructor to create new instances of Error.
func NewError(description string) Error {
	return Error{description}
}

// Error function to implement the error interface
// in case we ever get interface typed events.
func (err Error) Error() string {
	return err.err
}
