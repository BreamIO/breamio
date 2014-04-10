package beenleigh

import (
	"github.com/maxnordlund/breamio/briee"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
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

	go func() {
		bl.ListenAndServe()
		close(done)
	}()
	
	time.Sleep(time.Second)

	Convey("Some its events should be subscribed to", t, func() {
		t.Log(myEE)
		So(myEE.subscribedTo("shutdown"), ShouldEqual, true)
	})

	Convey("And events recieved handeled", t, func() {
		Convey("Returns when recieving a\"shutdown\" event", func() {
			myEE.pubsubs["shutdown"].(chan struct{}) <- struct{}{}
			_, ok := <-done
			So(ok, ShouldNotEqual, true)
		})
	})
}