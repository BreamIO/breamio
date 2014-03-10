package gorgonzola_test

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"github.com/maxnordlund/breamio/gorgonzola"
)

func TestGazeCreate(t *testing.T) {
	tracker, err := gorgonzola.GetDriver("gobii").Create()
	Convey("Result should be a tracker", t, func() {
		So(tracker, ShouldNotBeNil)
	})
	Convey("And the error should be nil.", t, func() {
		So(err, ShouldBeNil)
	})
}

func TestGazeList(t *testing.T) {
	driver := gorgonzola.GetDriver("gobii")
	Convey("Should always return a list", t, func() {
		So(driver.List(), ShouldNotBeNil)
	})
}

func TestGazeCreateFromId(t *testing.T) {
	driver := gorgonzola.GetDriver("gobii")
	ids := driver.List()
	if len(ids) < 1 {
		t.Fatal("No trackers connected.")
		return
	}
	tracker, err := driver.CreateFromId(ids[0])
	Convey("Result should be a tracker", t, func() {
		So(tracker, ShouldNotBeNil)
	})
	Convey("And the error should be nil.", t, func() {
		So(err, ShouldBeNil)
	})
}

func TestGazeStream(t *testing.T) {
	tracker, err := gorgonzola.GetDriver("gobii").Create()
	if err != nil {
		t.Fatal(err)
	}
	etdatas, errors := tracker.Stream()
	Convey("Should not give nil channels", t, func() {
		So(etdatas, ShouldNotBeNil)
		So(errors, ShouldNotBeNil)
	})

	SkipConvey("Should not recieve a error first", t, func() {
		good := false
		select {
		case <-etdatas:
			good = true
		case err := <-errors:
			t.Log(err)
			good = false
		}
		So(good, ShouldEqual, true)
	})

	SkipConvey("Closing during Stream should result in end of stream", t, func() {
		tracker.Close()
		<-etdatas //Value in pipeline
		_, ok := <-etdatas
		So(ok, ShouldEqual, false)
	})
}

func TestGazeCalibration(t *testing.T) {
	//tracker, _ := gorgonzola.GetDriver("mock").Create()
	Convey("Calibrating a GazeTracker eat all points on channel", t, nil)
}

func TestGazeIsCalibrated(t *testing.T) {
	//tracker, _ := gorgonzola.GetDriver("mock").Create()
	Convey("A GazeTracker should be calibrated after been given ~5 points", t, nil)
}
