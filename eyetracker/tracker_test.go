package eyetracker_test

import (
	"fmt"

	"github.com/maxnordlund/breamio/eyetracker"
	"github.com/maxnordlund/breamio/eyetracker/mock"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func ExampleUsage() {
	mocker := eyetracker.GetDriver("mock")
	if mocker == nil {
		fmt.Println("No Mock driver installed! :(")
		return
	}
	tracker, err := mocker.CreateFromId("constant")
	if err != nil {
		fmt.Println("No constant \"mocktracker\" implementated! :(")
		return
	}

	tracker.Connect()
	defer tracker.Close()
	points, errs := tracker.Stream()
	select {
	case p := <-points:
		fmt.Printf("(%0.2f, %0.2f)\n", p.Filtered.X(), p.Filtered.Y())
	case err = <-errs:
		fmt.Println("Error:", err)
	}
	// Output: (0.50, 0.50)
}

func TestRegisterDriver(t *testing.T) {
	Convey("Reregistration is not allowed", t, func() {
		So(eyetracker.RegisterDriver("mock", new(mock.MockDriver)), ShouldNotBeNil)
	})

	Convey("But new registrations should be allowed", t, func() {
		So(eyetracker.RegisterDriver("mock2", new(mock.MockDriver)), ShouldBeNil)
	})

	Convey("But only if the driver is not nil", t, func() {
		So(eyetracker.RegisterDriver("mock3", nil), ShouldNotBeNil)
	})
}

func TestList(t *testing.T) {
	Convey("List should not return nil", t, func() {
		So(eyetracker.List(), ShouldNotBeNil)
	})

	Convey("List should contain at least two elements", t, func() {
		So(len(eyetracker.List()), ShouldBeGreaterThanOrEqualTo, 2)
	})
}
