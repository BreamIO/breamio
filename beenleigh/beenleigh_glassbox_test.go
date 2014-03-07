package beenleigh

import (
	"testing"
	"time"
	. "github.com/smartystreets/goconvey/convey"
	
	"github.com/maxnordlund/breamio/aioli"
	"github.com/maxnordlund/breamio/gorgonzola"
)

func TestNewTrackerEvent(t *testing.T) {
	t.Skip("Does not work due to race conditions.")
	/*
		Due to the very asynchronous communication method used, we have no clue if or when stuff is done.
		I can not reliably test if this method does what it is supposed to since that means
			1. Check that a event emitter was created.
				This can almost be done.
				Code for it can be found below, but notice also the small sleep.
				Without it nothing would work. This is not a reliable test.
				
			2. Check that a tracker is created
				Again, almost.
				given the emitter, I can ask it for a ETData subscription.
				If the tracker is running, I should get events.
				Unfortunately, I can't ask it for one message, so it will spam the console with "dropped" messages.
				
		Oh, so you say I am doing to much in one method?
		Well.. Then what should I do.
		Creating a eye tracker is trivial thanks to code and tests in Gorgonzola.
		Creating a EventEmitter is trivial thanks to code and tests in Briee.
		
		So the real test should be, DOES it create them and what does it do with them.
		And as argued above, this is f*cking impossible to test.
		
		TL;DR Fuck Event emitters and testing.
	*/
	bl := New().(*breamLogic)
	Convey("Given a new:tracker event, a new tracker on a new emitter should be started.", t, func() {
		bl.onNewTrackerEvent(TrackerSpec{"mock", "constant", 1})
		time.Sleep(1*time.Millisecond)
		newEE := bl.MainIOManager().(*aioli.BasicIOManager).EEMap[1]
		So(newEE, ShouldNotBeNil)
		etEvents := newEE.Subscribe("tracker:etdata", &gorgonzola.ETData{}).(<-chan *gorgonzola.ETData)
		select {
			case data := <-etEvents:
				So(data, ShouldResemble, &gorgonzola.ETData{gorgonzola.Point2D{0.5, 0.5}, time.Now()})
				
			case <- time.After(time.Millisecond):
				t.Error("Timed out.")
		}
	})
	bl.Close()
}