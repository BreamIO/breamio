package mock_test

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"

	"github.com/maxnordlund/breamio/gorgonzola"
	_ "github.com/maxnordlund/breamio/gorgonzola/mock"
)

func TestCreate(t *testing.T) {
	tracker, err := gorgonzola.GetDriver("mock").Create()
	Convey("Result should be a tracker", t, func() {
		So(tracker, ShouldNotBeNil)
	})
	Convey("And the error should be nil.", t, func() {
		So(err, ShouldBeNil)
	})
}

func TestList(t *testing.T) {
	driver := gorgonzola.GetDriver("mock")
	Convey("Should always return a list", t, func() {
		So(driver.List(), ShouldNotBeNil)
	})
}

func TestCreateFromId(t *testing.T) {
	driver := gorgonzola.GetDriver("mock")
	for _, id := range driver.List() {
		tracker, err := driver.CreateFromId(id)
		Convey("(" + id + ") Result should be a tracker", t, func() {
			So(tracker, ShouldNotBeNil)
		})
		Convey("(" + id + ") And the error should be nil.", t, func() {
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
	driver := gorgonzola.GetDriver("mock")
	tracker, _ := driver.CreateFromId("constant")
	tracker.Connect()
	Convey("Value from Constant tracker should be constant.", t, func() {
		stream, _ := tracker.Stream()
		first := <-stream
		for i:=0; i < 20; i++ {
			data := <-stream
			So(data.Filtered.X(), ShouldResemble, first.Filtered.X())
			So(data.Filtered.Y(), ShouldResemble, first.Filtered.Y())
		}
	})
}

func TestStream(t *testing.T) {
	tracker, _ := gorgonzola.GetDriver("mock").Create()
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
	mee := &mockEmitter{make(map[string]interface{}), make(map[string]bool)}
	tracker, _ := gorgonzola.GetDriver("mock").Create()
	tracker.Connect()
	Convey("Link should set up some event handlers", t, func() {
		go tracker.Link(mee)
		time.Sleep(1*time.Second)
		So(mee.pubsubs["tracker:etdata"], ShouldNotBeNil)
		Convey("And register as publisher of the answers", func() {
			So(mee.pubsubs["tracker:calibrate:next"], ShouldNotBeNil)
			So(mee.pubsubs["tracker:calibrate:end"], ShouldNotBeNil)
			So(mee.pubsubs["tracker:validate:start"], ShouldNotBeNil)
			So(mee.pubsubs["tracker:validate:next"], ShouldNotBeNil)
			So(mee.pubsubs["tracker:validate:end"], ShouldNotBeNil)
		})
		
		Convey("tracker:calibrate:start", func() {
			So(mee.pubsubs["tracker:calibrate:start"], ShouldNotBeNil)
			mee.Dispatch("tracker:calibrate:start", struct{}{})
			So(<-mee.pubsubs["tracker:calibrate:next"].(chan struct{}), ShouldResemble, struct{}{})
		})
		
		Convey("tracker:calibrate:add", func(){
			So(mee.pubsubs["tracker:calibrate:add"], ShouldNotBeNil)
			addCh := mee.Publish("tracker:calibrate:add", gorgonzola.Point2D{}).(chan<- gorgonzola.Point2D)
			addCh <- gorgonzola.Point2D{0.1,0.1}
			So(<-mee.pubsubs["tracker:calibrate:next"].(chan struct{}), ShouldResemble, struct{}{})
			addCh <- gorgonzola.Point2D{0.9,0.1}
			So(<-mee.pubsubs["tracker:calibrate:next"].(chan struct{}), ShouldResemble, struct{}{})
			addCh <- gorgonzola.Point2D{0.1,0.9}
			So(<-mee.pubsubs["tracker:calibrate:next"].(chan struct{}), ShouldResemble, struct{}{})
			addCh <- gorgonzola.Point2D{0.9,0.9}
			So(<-mee.pubsubs["tracker:calibrate:next"].(chan struct{}), ShouldResemble, struct{}{})
			addCh <- gorgonzola.Point2D{0.5,0.5}
			So(<-mee.pubsubs["tracker:calibrate:end"].(chan struct{}), ShouldResemble, struct{}{})
			So(<-mee.pubsubs["tracker:validate:start"].(chan struct{}), ShouldResemble, struct{}{})
		})
		
		Convey("tracker:validate:start", func() {
			So(mee.pubsubs["tracker:validate:start"], ShouldNotBeNil)
			mee.Dispatch("tracker:validate:start", struct{}{})
			So(<-mee.pubsubs["tracker:validate:next"].(chan struct{}), ShouldResemble, struct{}{})
		})
		
		Convey("tracker:validate:add", func(){
			So(mee.pubsubs["tracker:validate:add"], ShouldNotBeNil)
			addCh := mee.Publish("tracker:validate:add", gorgonzola.Point2D{}).(chan<- gorgonzola.Point2D)
			addCh <- gorgonzola.Point2D{0.1,0.1}
			So(<-mee.pubsubs["tracker:validate:next"].(chan struct{}), ShouldResemble, struct{}{})
			addCh <- gorgonzola.Point2D{0.9,0.1}
			So(<-mee.pubsubs["tracker:validate:next"].(chan struct{}), ShouldResemble, struct{}{})
			addCh <- gorgonzola.Point2D{0.1,0.9}
			So(<-mee.pubsubs["tracker:validate:next"].(chan struct{}), ShouldResemble, struct{}{})
			addCh <- gorgonzola.Point2D{0.9,0.9}
			So(<-mee.pubsubs["tracker:validate:next"].(chan struct{}), ShouldResemble, struct{}{})
			addCh <- gorgonzola.Point2D{0.5,0.5}
			So(<-mee.pubsubs["tracker:validate:end"].(chan float64), ShouldResemble, float64(0.05))
		})
		
		Convey("Closing should shut them all down.", func() {
			tracker.Close()
			time.Sleep(1*time.Second)
			So(mee.unsubscribed["tracker:calibrate:start"], ShouldEqual, true)
			So(mee.unsubscribed["tracker:calibrate:add"], ShouldEqual, true)
			So(mee.unsubscribed["tracker:validate:start"], ShouldEqual, true)
			So(mee.unsubscribed["tracker:validate:add"], ShouldEqual, true)
		})
	})
}

type mockEmitter struct {
	pubsubs map[string]interface{}
	unsubscribed map[string]bool
}

func (m *mockEmitter) create(eventID string, typ interface{}) {
	if _, ok := m.pubsubs[eventID]; ok {
		return
	} else {
		switch typ.(type) {
			case *gorgonzola.ETData:
				m.pubsubs[eventID] = make(chan *gorgonzola.ETData)
			case gorgonzola.Point2D:
				m.pubsubs[eventID] = make(chan gorgonzola.Point2D)
			case struct{}:
				m.pubsubs[eventID] = make(chan struct{})
			case float64:
				m.pubsubs[eventID] = make(chan float64)
		}
	}
}

func (m *mockEmitter) Publish(eventID string, typ interface{}) interface{} {
	m.create(eventID, typ)
	switch typ.(type) {
		case *gorgonzola.ETData: return (chan<- *gorgonzola.ETData)(m.pubsubs[eventID].(chan *gorgonzola.ETData))
		case gorgonzola.Point2D: return (chan<- gorgonzola.Point2D)(m.pubsubs[eventID].(chan gorgonzola.Point2D))
		case struct{}:           return (chan<- struct{})(m.pubsubs[eventID].(chan struct{}))
		case float64:            return (chan<- float64)(m.pubsubs[eventID].(chan float64))
	}
	return m.pubsubs[eventID]
}

func (m *mockEmitter) Subscribe(eventID string, typ interface{}) interface{} {
	m.create(eventID, typ)
	switch typ.(type) {
		case *gorgonzola.ETData: return (<-chan *gorgonzola.ETData)(m.pubsubs[eventID].(chan *gorgonzola.ETData))
		case gorgonzola.Point2D: return (<-chan gorgonzola.Point2D)(m.pubsubs[eventID].(chan gorgonzola.Point2D))
		case struct{}:           return (<-chan struct{})(m.pubsubs[eventID].(chan struct{}))
		case float64:
	}
	return (m.pubsubs[eventID])
}

func (m *mockEmitter) Dispatch(eventID string, v interface{}) {
	if m.pubsubs[eventID] != nil {
		m.pubsubs[eventID].(chan struct{}) <- v.(struct{})
	}
}

func (m *mockEmitter) Unsubscribe(eventID string, typ interface{}) error {
	m.unsubscribed[eventID] = true
	return nil
}
