package gorgonzola

import (
	"github.com/maxnordlund/breamio/comte"
	"github.com/maxnordlund/breamio/module"
	//"encoding/json"
	"errors"
	"strings"
	"time"

	bl "github.com/maxnordlund/breamio/beenleigh"
)

func init() {
	gorgonzola := &GorgonzolaRun{closing: make(chan struct{})}
	gorgonzola.logger = bl.NewLogger(gorgonzola)
	bl.Register(gorgonzola)
}

var trackers = make(map[Tracker]Metadata)

//
func RemoveTracker(tracker Tracker) {
	if _, ok := trackers[tracker]; ok {
		delete(trackers, tracker)
	}
}

// Returns a map of all known running trackers.
// Only tracks trackers created using the new:tracker event.
func Trackers() map[Tracker]Metadata {
	return trackers
}

type GorgonzolaRun struct {
	logger  module.Logger
	closing chan struct{}
}

func (GorgonzolaRun) String() string {
	return "EyeTracking"
}

type Config []Metadata

func (GorgonzolaRun) Config() comte.ConfigSection {
	return make(Config, 0, 10)
}

func (gr GorgonzolaRun) New(c module.Constructor) module.Module {
	tracker, err := CreateFromURI(c,
		c.Parameters["URI"].(string))
	if err != nil {
		gr.logger.Printf("Could not create new tracker with uri %s: %s", c.Parameters["URI"].(string), err)
		panic(err)
	}
	err = tracker.Connect()
	if err != nil {
		gr.logger.Println("Unable to connect to tracker:", err)
		panic(err)
	}

	bl.RunModule(c.Logic, c.Parameters["Emitter"], tracker)

	return tracker
}

func (gr GorgonzolaRun) Run(logic bl.Logic) {
	var conf Config
	comte.Section(gr.String(), &conf)
	for _, md := range conf {
		gr.logger.Printf("Creating %s on %d from config", md.URI, md.Emitter)
		if tracker, err := gr.NewTracker(logic, md); err != nil {
			logic.RootEmitter().Dispatch("gorgonzola:error", err)
		}
	}

	bl.RunFactory(logic, gr)

}

func (gr GorgonzolaRun) Close() error {
	close(gr.closing)
	return nil
}

func (gr GorgonzolaRun) NewTracker(logic bl.Logic, event Metadata) (Tracker, error) {
	tracker, err := CreateFromURI(event.URI)
	if err != nil {
		gr.logger.Printf("Could not create new tracker with uri %s: %s", event.URI, err)
		return err
	}
	err = tracker.Connect()
	if err != nil {
		gr.logger.Println("Unable to connect to tracker:", err)
		return err
	}

	ee := logic.CreateEmitter(event.Emitter)
	go tracker.Link(ee)

	gr.logger.Printf("Created a new tracker with uri %s on EE %d.\n", event.URI, event.Emitter)
	trackers[tracker] = event
	return nil
}

// Creates a tracker using a URI.
// The URI is on the form <driver>://<id>.
// The driver part is used to find a registered driver in the tracker driver table.
// The id part is used as a argument for the selected drivers CreateFromId method.
func CreateFromURI(c module.Constructor, uri string) (Tracker, error) {
	split := strings.SplitN(uri, "://", 2)
	if len(split) < 2 {
		return nil, errors.New("Malformed URI.")
	}
	typ, id := split[0], split[1]
	driver := GetDriver(typ)
	if driver == nil {
		return nil, errors.New("No such driver.")
	}
	return GetDriver(typ).CreateFromId(c, id)
}

// A interface specifying something that can deliver
// x and y coordinates in Cartesian space.
type XYer interface {
	X() float64
	Y() float64
}

type XYZer interface {
	XYer
	Z() float64
}

func ToPoint2D(in XYer) (out *Point2D) {
	out, ok := in.(*Point2D)
	if ok {
		return
	}
	out = new(Point2D)
	out.Xf = in.X()
	out.Yf = in.Y()
	return
}

func ToPoint3D(in XYZer) (out *Point3D) {
	out, ok := in.(*Point3D)
	if ok {
		return
	}
	out = new(Point3D)
	out.Xf = in.X()
	out.Yf = in.Y()
	out.Zf = in.Z()
	return
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

type Point3D struct {
	Xf, Yf, Zf float64
}

// Returns the X part of the Point3D
func (p Point3D) X() float64 {
	return p.Xf
}

// Returns the Y part of the Point3D
func (p Point3D) Y() float64 {
	return p.Yf
}

// Returns the Z part of the Point3D
func (p Point3D) Z() float64 {
	return p.Zf
}

func Filter(left, right XYer) XYer {
	return Point2D{(left.X() + right.X()) / 2, (left.Y() + right.Y()) / 2}
}

type ETData struct {
	Filtered  Point2D
	LeftGaze  *Point2D `json:",omitempty"`
	RightGaze *Point2D `json:",omitempty"`
	LeftEye   *Point3D `json:",omitempty"`
	RightEye  *Point3D `json:",omitempty"`
	Timestamp time.Time
}

type Metadata struct {
	Emitter int
	URI     string
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
