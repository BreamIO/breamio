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
	Convey("Call any registered runners close methods", t, func() {
		runners = runners[:0]
		closed := false
		Register(NewRunHandler(func (l Logic, cch <-chan struct{} ) {
			<-cch
			closed = true
		}))
		bl = newBL(briee.New)
		go bl.ListenAndServe()
		bl.Close()
		time.Sleep(time.Second)
		So(closed, ShouldEqual, true)
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
	
	runners = runners[:0]
	first, second := false, false
	Register(NewRunHandler(func(l Logic, cch <-chan struct{}) { first = true }))
	Register(NewRunHandler(func(l Logic, cch <-chan struct{}) { second = true }))

	go func() {
		bl.ListenAndServe()
		close(done)
	}()
	
	time.Sleep(time.Second)
	
	Convey("Starts all registered runners.", t, func(){
		So(first, ShouldEqual, true)
		So(second, ShouldEqual, true)
	})

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

func TestCreateEmitter(t *testing.T) {
	bl := New(briee.New)
	
	//Nil test
	ee1 := bl.CreateEmitter(1)
	if ee1 == nil {
		t.Error("Created emitter 1 was nil.")
	}
	eeNeg1 := bl.CreateEmitter(-1)
	if eeNeg1 == nil {
		t.Error("Created emitter -1 was nil.")
	}
	
	//Same test
	
	if ee1_2 := bl.CreateEmitter(1); ee1 != ee1_2 {
		t.Error("Emitter was overwritten with new.")
	}
	
	//!Same test
	if ee2 := bl.CreateEmitter(2); ee1 == ee2 && ee2 != nil {
		t.Error("Emitter 1 was reused as emitter 2.")
	}
	if ee17 := bl.CreateEmitter(17); ee1 == ee17 && ee17 != nil {
		t.Error("Emitter 1 was reused as emitter 2.")
	}
	if ee42 := bl.CreateEmitter(42); ee1 == ee42 && ee42 != nil {
		t.Error("Emitter 1 was reused as emitter 2.")
	}
	
	
}

func TestEmitterLookup(t *testing.T) {
	bl := New(briee.New)
	
	//Create some emitters to check against.
	ee1 := bl.CreateEmitter(1)
	ee2 := bl.CreateEmitter(2)
	ee18 := bl.CreateEmitter(18)
	
	if ee1_2, err := bl.EmitterLookup(1); ee1 != ee1_2 || err != nil {
		t.Error("Same emitter 1 was not returned")
	}
	
	if ee2_2, err := bl.EmitterLookup(2); ee2 != ee2_2 || err != nil {
		t.Error("Same emitter 2 was not returned")
	}
	
	if ee18_2, err := bl.EmitterLookup(18); ee18 != ee18_2 || err != nil {
		t.Error("Same emitter 18 was not returned")
	}
	
	//Check consistency
	for i:=0; i< 100; i++ {
		if ee_temp, err := bl.EmitterLookup(1); ee1 != ee_temp || err != nil {
			t.Fatal("Consitency problem with emitter 1.")
		}
		if ee_temp, err := bl.EmitterLookup(2); ee2 != ee_temp || err != nil {
			t.Fatal("Consitency problem with emitter 1.")
		}
		if ee_temp, err := bl.EmitterLookup(18); ee18 != ee_temp || err != nil {
			t.Fatal("Consitency problem with emitter 1.")
		}
	}
	
	if _, err := bl.EmitterLookup(4711); err == nil {
		t.Error("EmitterLookup created new emitter.")
	}
	
}
