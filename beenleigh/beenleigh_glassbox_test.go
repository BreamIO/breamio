package beenleigh

import (
	"github.com/maxnordlund/breamio/briee"
	_ "github.com/maxnordlund/breamio/gorgonzola/mock"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestClose(t *testing.T) {
	bl := newBL(briee.New)
	Convey("Should close the internal closer channel", t, func() {
		go bl.Close()
		_, ok := <-bl.closer
		So(ok, ShouldNotEqual, true)
	})
	Convey("Return value should be nil", t, func() {
		bl = newBL(briee.New)
		So(bl.Close(), ShouldBeNil)
	})
}

func TestRootEmitter(t *testing.T) {
	bl := newBL(briee.New)
	Convey("A first root event emitter should be running.", t, func() {
		So(bl.RootEmitter(), ShouldNotEqual, nil)
	})
}

func TestListenAndServe(t *testing.T) {
	myEE := newMockEmitter()
	done := make(chan struct{})
	bl := newBL(func() briee.EventEmitter { return myEE })
	ioman := newMockIOManager()

	go func() {
		bl.ListenAndServe()
		close(done)
	}()

	Convey("Some its events should be subscribed to", t, func() {
		t.Log(myEE)
		So(myEE.subscribedTo("shutdown"), ShouldEqual, true)
	})

	Convey("The IOManager should be started.", t, func() {
		So(myIOManager.started, ShouldEqual, true)
	})

	Convey("And events recieved handeled", t, func() {
		Convey("Returns when recieving a\"shutdown\" event", func() {
			myEE.pubsubs["shutdown"].(chan struct{}) <- struct{}{}
			_, ok := <-done
			So(ok, ShouldNotEqual, true)
		})
	})
}

// func TestOnNewTrackerEvent(t *testing.T) {
	// myEE := newMockEmitter()
	// myIOManager := newMockIOManager()
	// bl := newBL(func() briee.EventEmitter { return myEE }, myIOManager)
	// onNewTrackerEvent(bl, Spec{1, "mock://constant"})
	// Convey("Creates a new EE and adds it to IOManager", t, func() {
		// So(myIOManager.ees[1], ShouldEqual, myEE)
	// })
	// SkipConvey("Creates a Tracker from specification and connects it to EE", t, func() {
		// So(onNewTrackerEvent(bl, Spec{2, "mock://ShouldNotExist"}), ShouldNotBeNil)
	// })
// }

/*
func TestNewStatisticsEvent(t *testing.T) {
	Convey("new:statistics events should spawn a new Stilton (name pending) instance for specified tracker.", t, nil)
}*/
