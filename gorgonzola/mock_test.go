package gorgonzola_test

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"

	"github.com/maxnordlund/breamio/gorgonzola"
)

func TestMockCreate(t *testing.T) {
	tracker, err := gorgonzola.GetDriver("mock").Create()
	Convey("Result should be a tracker", t, func() {
		So(tracker, ShouldNotBeNil)
	})
	Convey("And the error should be nil.", t, func() {
		So(err, ShouldBeNil)
	})
}

func TestMockList(t *testing.T) {
	driver := gorgonzola.GetDriver("mock")
	Convey("Should always return a list", t, func() {
		So(driver.List(), ShouldNotBeNil)
	})
}

func TestMockCreateFromId(t *testing.T) {
	driver := gorgonzola.GetDriver("mock")
	id := driver.List()[0]
	tracker, err := driver.CreateFromId(id)
	Convey("Result should be a tracker", t, func() {
		So(tracker, ShouldNotBeNil)
	})
	Convey("And the error should be nil.", t, func() {
		So(err, ShouldBeNil)
	})
	Convey("Creating from non-existing id should give an error", t, func() {
		t2, err := driver.CreateFromId("This ID Should Not Exist")
		So(t2, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})
}

func TestMockStream(t *testing.T) {
	tracker, _ := gorgonzola.GetDriver("mock").Create()
	etdatas, errors := tracker.Stream()
	Convey("Should not give nil channels", t, func() {
		So(etdatas, ShouldNotBeNil)
		So(errors, ShouldNotBeNil)
	})

	Convey("Should not recieve a error first", t, func() {
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

	Convey("Closing during Stream should result in end of stream", t, func() {
		tracker.Close()
		<-etdatas //Value in pipeline
		_, ok := <-etdatas
		So(ok, ShouldEqual, false)
	})
}

func TestMockCalibration(t *testing.T) {
	tracker, _ := gorgonzola.GetDriver("mock").Create()
	Convey("Calibrating a MockTracker should not work", t, func() {
		errs := make(chan error, 1)
		tracker.Calibrate(nil, errs)
		So(<-errs, ShouldNotBeNil)
	})
}

func TestMockIsCalibrated(t *testing.T) {
	tracker, _ := gorgonzola.GetDriver("mock").Create()
	Convey("A MockTracker should never be calibrated", t, func() {
		errs := make(chan error, 1)
		tracker.Calibrate(nil, errs)
		So(tracker.IsCalibrated(), ShouldEqual, false)

		tracker.Calibrate(nil, make(chan error, 1))
		So(tracker.IsCalibrated(), ShouldEqual, false)
	})
}
