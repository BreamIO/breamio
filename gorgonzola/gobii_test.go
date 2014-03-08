package gorgonzola_test

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
	
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
	id := driver.List()[0]
	tracker, err := driver.CreateFromId(id)
	Convey("Result should be a tracker", t, func() {
		So(tracker, ShouldNotBeNil)
	})
	Convey("And the error should be nil.", t, func() {
		So(err, ShouldBeNil)
	})
}

func TestGazeStream(t *testing.T) {
	tracker, _ := gorgonzola.GetDriver("gobii").Create()
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
}

