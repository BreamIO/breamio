package gorgonzola

import (
	"testing"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFilter(t *testing.T) {
	Convey("Filter should not return nil", t, func() {
		So(filter(point2D{}, point2D{}), ShouldNotBeNil)
	})
	
	Convey("Filter should average the points given in both axis", t ,func(){
		So(filter(point2D{-1, -1}, point2D{1, 1}), ShouldResemble, point2D{})
		So(filter(point2D{0.5, 0.5}, point2D{0.5, 0.5}), ShouldResemble, point2D{0.5, 0.5})
	})
}

func TestLink(t *testing.T) {
	SkipConvey("Link should connect a tracker to a event emitter", t, func(){
		SkipConvey("Tracker should publish ETData on \"tracker:etdata\" event", nil)
	})
}

func TestListenAndServe(t *testing.T) {
	dataCh, errCh, pubCh := make(chan *ETData, 1), make(chan error, 1), make(chan *ETData, 1)
	Convey("A error on error channel should make it return that error", t, func() {
		original_error := NotImplementedError("HaHa!")
		errCh <- original_error
		err := listenAndServe(dataCh, errCh, pubCh)
		So(err, ShouldEqual, original_error)
	})
	
	Convey("A ETData on the data channel should make it publish that on the publish channel", t, func() {
		original_data := new(ETData)
		dataCh <- original_data
		go listenAndServe(dataCh, errCh, pubCh)
		So(<-pubCh, ShouldEqual, original_data)
		errCh <- NotImplementedError("Closing time.") //Lets avoid leaking go routines in the tests, shall we?
	})
	
	Convey("Adding multiple ETData on data channel without published read should not create deadlock.", t, func(){
		storageArea := error(nil)
		done := make(chan struct{})
		go func() {
			storageArea = listenAndServe(dataCh, errCh ,pubCh)
			close(done)
		}()
		dataCh <- new(ETData)
		dataCh <- new(ETData)
		dataCh <- new(ETData)
		errCh <- NotImplementedError("Deadlock test.")
		<-done
		So(storageArea, ShouldNotBeNil)
		
	})
	
	Convey("Closing data channel should make it stop, returning nil", t, func() {
		var storageArea error = NotImplementedError("") //Testing error. Expected to disappear.
		done := make(chan struct{})
		go func() {
			storageArea = listenAndServe(dataCh, errCh ,pubCh)
			close(done)
		}()
		close(dataCh)
		<-done
		So(storageArea, ShouldBeNil)
	})
	
}