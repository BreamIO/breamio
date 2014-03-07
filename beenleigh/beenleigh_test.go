package beenleigh_test

import(
	"testing"
	"time"
	"reflect"
	"github.com/maxnordlund/breamio/beenleigh"
	. "github.com/smartystreets/goconvey/convey"
)

/*
	This is really wrong.
	It should not have to be this way.
	
	If anyone can find a better way to test this,
	Feel free to do it, because I can't do it.
*/
func TestClose(t *testing.T) {
	bl := beenleigh.New()
	
	done := make(chan struct{})
	go func() {
		bl.ListenAndServe()
		close(done)
	}()
	Convey("Return value always be nil", t, func() {
		So(bl.Close(), ShouldBeNil)
	})
	
	select {
	case <-done:
		//Closed as expected.
	case <-time.After(1*time.Second): //Timeout to allow ListenAndServe to actually close.
		t.Fail()
	}
}

func TestRootEmitter(t *testing.T) {
	bl := beenleigh.New()
	Convey("A first root event emitter should be running.", t, func() {
		So(bl.RootEmitter(), ShouldNotEqual, nil)
	})
}

func TestMainIOManager(t *testing.T) {
	bl := beenleigh.New()
	Convey("And a IOManager should be available.", t, func() {
		So(bl.MainIOManager(), ShouldNotEqual, nil)
	})
}

func TestListenAndServe(t *testing.T) {
	bl := beenleigh.New()
	go bl.ListenAndServe()
	time.Sleep(time.Millisecond) //Let ListenAndServe run and do its stuff.
	Convey("Some events should be subscribed to", t, func(){
		Convey("Like \"new:tracker\"", func(){
			typ, err := bl.RootEmitter().TypeOf("new:tracker")
			if err != nil {
				t.Error("Could not get type of new:tracker:", err)
			}
			So(typ, ShouldEqual, reflect.TypeOf(beenleigh.TrackerSpec{}))
		})
		Convey("And \"new:statistics\"", nil)
		Convey("Don't forget \"shutdown\"", func(){
			typ, err := bl.RootEmitter().TypeOf("shutdown") 
			if err != nil {
				t.Error("Could not get type of new:tracker:", err)
			}
			So(typ, ShouldEqual, reflect.TypeOf(beenleigh.TrackerSpec{}))
		})
	})
	bl.Close()
}

func TestShutdownEvent(t *testing.T) {
	Convey("A Shutdown event recieved should make it stop all EE and trackers", t, nil)
}

func TestNewStatisticsEvent(t *testing.T) {
	Convey("new:statistics events should spawn a new Stilton (name pending) instance for specified tracker.", t, nil)
}

