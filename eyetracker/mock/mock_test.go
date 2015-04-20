package mock_test

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"

	"github.com/maxnordlund/breamio/briee"
	"github.com/maxnordlund/breamio/eyetracker"
	_ "github.com/maxnordlund/breamio/eyetracker/mock"
	gt "github.com/maxnordlund/breamio/eyetracker/testing"
)

func TestCreate(t *testing.T) {
	tracker, err := eyetracker.GetDriver("mock").Create()
	Convey("Result should be a tracker", t, func() {
		So(tracker, ShouldNotBeNil)
	})
	Convey("And the error should be nil.", t, func() {
		So(err, ShouldBeNil)
	})
}

func TestList(t *testing.T) {
	driver := eyetracker.GetDriver("mock")
	Convey("Should always return a list", t, func() {
		So(driver.List(), ShouldNotBeNil)
	})
}

func TestCreateFromId(t *testing.T) {
	driver := eyetracker.GetDriver("mock")
	for _, id := range driver.List() {
		tracker, err := driver.CreateFromId(id)
		Convey("("+id+") Result should be a tracker", t, func() {
			So(tracker, ShouldNotBeNil)
		})
		Convey("("+id+") And the error should be nil.", t, func() {
			So(err, ShouldBeNil)
		})
	}
	Convey("Creating from non-existing id should give an error", t, func() {
		t2, err := driver.CreateFromId("This ID Should Not Exist")
		So(t2, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})
}

func TestConstant(t *testing.T) {
	driver := eyetracker.GetDriver("mock")
	tracker, _ := driver.CreateFromId("constant")
	tracker.Connect()
	Convey("Value from Constant tracker should be constant.", t, func() {
		stream, _ := tracker.Stream()
		first := <-stream
		for i := 0; i < 20; i++ {
			data := <-stream
			So(data.Filtered.X(), ShouldResemble, first.Filtered.X())
			So(data.Filtered.Y(), ShouldResemble, first.Filtered.Y())
		}
	})
}

func TestStream(t *testing.T) {
	tracker, _ := eyetracker.GetDriver("mock").Create()
	tracker.Connect()
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

func TestLink(t *testing.T) {
	mee := &gt.MockEmitter{make(map[string]interface{}), make(map[string]bool)}
	tracker, _ := eyetracker.GetDriver("mock").Create()
	tracker.Connect()
	Convey("Link should set up some event handlers", t, func() {
		go tracker.Link(mee)
		time.Sleep(1 * time.Second)
		So(mee.Pubsubs["tracker:etdata"], ShouldNotBeNil)
		Convey("And register as publisher of the answers", func() {
			So(mee.Pubsubs["tracker:calibrate:start"], ShouldNotBeNil)
			So(mee.Pubsubs["tracker:calibrate:next"], ShouldNotBeNil)
			So(mee.Pubsubs["tracker:calibrate:end"], ShouldNotBeNil)
			So(mee.Pubsubs["tracker:validate:start"], ShouldNotBeNil)
			So(mee.Pubsubs["tracker:validate:next"], ShouldNotBeNil)
			So(mee.Pubsubs["tracker:validate:end"], ShouldNotBeNil)
		})
	})
}

func TestClose(t *testing.T) {
	mee := &gt.MockEmitter{make(map[string]interface{}), make(map[string]bool)}
	tracker, _ := eyetracker.GetDriver("mock").Create()
	tracker.Connect()
	go tracker.Link(mee)
	SkipConvey("Closing should shut down all subscriptions", t, func() {
		tracker.Close()
		t.Log("-1")
		time.Sleep(2 * time.Second)
		t.Log("0")
		So(mee.Unsubscribed["tracker:calibrate:start"], ShouldEqual, true)
		t.Log("1")
		So(mee.Unsubscribed["tracker:calibrate:add"], ShouldEqual, true)
		t.Log("2")
		So(mee.Unsubscribed["tracker:validate:start"], ShouldEqual, true)
		t.Log("3")
		So(mee.Unsubscribed["tracker:validate:add"], ShouldEqual, true)
		t.Log("4")
	})
}

func TestCalibration(t *testing.T) {
	tracker, _ := eyetracker.GetDriver("mock").Create()
	tracker.Connect()
	ee := briee.New()
	tracker.Link(ee)

	calib_nextCh := ee.Subscribe("tracker:calibrate:next", struct{}{}).(<-chan struct{})
	calib_errorCh := ee.Subscribe("tracker:calibrate:error", eyetracker.NewError("")).(<-chan eyetracker.Error)
	calib_endCh := ee.Subscribe("tracker:calibrate:end", struct{}{}).(<-chan struct{})
	valid_startCh := ee.Subscribe("tracker:validate:start", struct{}{}).(<-chan struct{})
	valid_nextCh := ee.Subscribe("tracker:validate:next", struct{}{}).(<-chan struct{})
	valid_endCh := ee.Subscribe("tracker:validate:end", float64(0)).(<-chan float64)

	Convey("tracker:calibrate:start", t, func() {
		ee.Dispatch("tracker:calibrate:start", struct{}{})
		So(gt.CheckError(calib_nextCh, calib_errorCh), ShouldBeNil)
	})

	Convey("tracker:calibrate:add", t, func() {
		addCh := ee.Publish("tracker:calibrate:add", eyetracker.Point2D{}).(chan<- eyetracker.Point2D)
		defer close(addCh)

		addCh <- eyetracker.Point2D{0.1, 0.1}
		So(gt.CheckError(calib_nextCh, calib_errorCh), ShouldBeNil)

		addCh <- eyetracker.Point2D{0.9, 0.1}
		So(gt.CheckError(calib_nextCh, calib_errorCh), ShouldBeNil)

		addCh <- eyetracker.Point2D{0.1, 0.9}
		So(gt.CheckError(calib_nextCh, calib_errorCh), ShouldBeNil)

		addCh <- eyetracker.Point2D{0.9, 0.9}
		So(gt.CheckError(calib_nextCh, calib_errorCh), ShouldBeNil)

		addCh <- eyetracker.Point2D{0.5, 0.5}
		So(gt.CheckError(calib_endCh, calib_errorCh), ShouldBeNil)
		So(<-valid_startCh, ShouldResemble, struct{}{})
	})

	Convey("tracker:validate:start", t, func() {
		ee.Dispatch("tracker:validate:start", struct{}{})
		So(<-valid_nextCh, ShouldResemble, struct{}{})
	})

	Convey("tracker:validate:add", t, func() {
		addCh := ee.Publish("tracker:validate:add", eyetracker.Point2D{}).(chan<- eyetracker.Point2D)
		addCh <- eyetracker.Point2D{0.1, 0.1}
		So(<-valid_nextCh, ShouldResemble, struct{}{})
		addCh <- eyetracker.Point2D{0.9, 0.1}
		So(<-valid_nextCh, ShouldResemble, struct{}{})
		addCh <- eyetracker.Point2D{0.1, 0.9}
		So(<-valid_nextCh, ShouldResemble, struct{}{})
		addCh <- eyetracker.Point2D{0.9, 0.9}
		So(<-valid_nextCh, ShouldResemble, struct{}{})
		addCh <- eyetracker.Point2D{0.5, 0.5}
		So(<-valid_endCh, ShouldResemble, float64(0.05))
	})
}
