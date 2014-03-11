package beenleigh

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
	
	"github.com/maxnordlund/breamio/briee"
	"github.com/maxnordlund/breamio/gorgonzola"
)

func TestClose(t *testing.T) {
	bl := newBL(briee.New, nil)
	Convey("Should close the internal closer channel", t, func() {
		go bl.Close()
		_, ok := <-bl.closer
		So(ok, ShouldNotEqual, true)
	})
	Convey("Return value should be nil", t, func() {
		bl = newBL(briee.New, nil)
		So(bl.Close(), ShouldBeNil)
	})
}

func TestRootEmitter(t *testing.T) {
	bl := newBL(briee.New, nil)
	Convey("A first root event emitter should be running.", t, func() {
		So(bl.RootEmitter(), ShouldNotEqual, nil)
	})
}

func TestListenAndServe(t *testing.T) {
	myEE := newMockEmitter()
	myIOManager := newMockIOManager()
	done := make(chan struct{})
	bl := newBL(func() briee.EventEmitter {return myEE}, myIOManager)
	
	go func() {
		bl.ListenAndServe()
		close(done)
	}()
	
	Convey("Some its events should be subscribed to", t, func(){
		t.Log(myEE)
		So(myEE.subscribedTo("new"), ShouldEqual, true)
		So(myEE.subscribedTo("shutdown"), ShouldEqual, true)
	})
	
	Convey("The IOManager should be started.", t, func() {
		So(myIOManager.started, ShouldEqual, true)
	})
	
	Convey("And events recieved handeled", t, func() {
		Convey("Calls onNewTrackerEvent for \"new\" with type \"tracker\"", func() {
			done := make(chan struct{})
			bl.onNewTrackerEvent = func(spec Spec) error {
				close(done)
				return nil
			}
			myEE.pubsubs["new"].(chan Spec) <- Spec{"tracker", "", 0}
			_, ok := <-done
			So(ok, ShouldNotEqual, true)
			
		})
		Convey("Returns when recieving a\"shutdown\" event", func(){
			myEE.pubsubs["shutdown"].(chan struct{}) <- struct{}{}
			_, ok := <-done
			So(ok, ShouldNotEqual, true)
		})
	})
	
	Convey("And closes when asked to", t, func() {
		done := make(chan struct{})
		go func() {
			bl.ListenAndServe()
			close(done)
		}()
		close(bl.closer)
		_, ok := <-done
		So(ok, ShouldNotEqual, true)
	})
}

func TestOnNewTrackerEvent(t *testing.T) {
	myEE := newMockEmitter()
	myIOManager := newMockIOManager()
	gorgonzola.RegisterDriver("beenleigh_mock", &BLMockTrackerDriver{&gorgonzola.MockTracker{}, false, ""})
	bl := newBL(func() briee.EventEmitter {return myEE}, myIOManager)
	onNewTrackerEvent(bl, Spec{"tracker", "beenleigh_mock://test", 1})
	Convey("Creates a new EE and adds it to IOManager", t, func() {
		So(myIOManager.ees[1], ShouldEqual, myEE)
	})
	SkipConvey("Creates a Tracker from specification and connects it to EE", t, func(){
		So(onNewTrackerEvent(bl, Spec{"tracker", "beenleigh_mock://error", 2}), ShouldNotBeNil)
	})
}

/*
func TestNewStatisticsEvent(t *testing.T) {
	Convey("new:statistics events should spawn a new Stilton (name pending) instance for specified tracker.", t, nil)
}*/
