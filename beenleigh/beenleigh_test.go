package beenleigh_test

import(
	"testing"
	"reflect"
	bl "github.com/maxnordlund/breamio/beenleigh"
	. "github.com/smartystreets/goconvey/convey"
)

func TestOnStart(t *testing.T) {
	Convey("During start up of BL:", t, func() {
		Convey("A first root event emitter should be running.", func() {
			So(bl.RootEventEmitter(), ShouldNotEqual, nil)
		})
		Convey("And a IOManager should be available.", func() {
			So(bl.MainIOManager(), ShouldNotEqual, nil)
		})
		
		go bl.ListenAndServe()
		
		Convey("Some events should be subscribed to", func(){
			Convey("Like \"new:tracker\"", func(){
				typ, err := bl.RootEventEmitter().TypeOf("new:tracker")
				if err != nil {
					t.Fatal("Could not get type of new:tracker:", err)
				}
				So(typ, ShouldEqual, reflect.TypeOf(bl.TrackerSpec{}))
			})
			Convey("And \"new:statistics\"", nil)
			Convey("Don't forget \"shutdown\"", func(){
				typ, err := bl.RootEventEmitter().TypeOf("shutdown") 
				if err != nil {
					t.Fatal("Could not get type of new:tracker:", err)
				}
				So(typ, ShouldEqual, reflect.TypeOf(bl.TrackerSpec{}))
			})
		})
	})
}

func TestNewTrackerEvent(t *testing.T) {
	Convey("Given a new:tracker event, a new tracker on a new emitter should be started.", t, nil)
}

func TestShutdownEvent(t *testing.T) {
	Convey("A Shutdown event recieved should make it stop all EE and trackers", t, nil)
}

func TestNewStatisticsEvent(t *testing.T) {
	Convey("new:statistics events should spawn a new Stilton (name pending) instance for specified tracker.", t, nil)
}

