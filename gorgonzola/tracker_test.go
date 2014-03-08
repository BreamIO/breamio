package gorgonzola_test

import (
	"fmt"

	"github.com/maxnordlund/breamio/gorgonzola"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func ExampleUsage() {
	mock := gorgonzola.GetDriver("mock")
	if mock == nil {
		fmt.Println("No Mock driver installed! :(")
		return
	}
	tracker, err := mock.CreateFromId("constant")
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
		So(gorgonzola.RegisterDriver("mock", new(gorgonzola.MockDriver)), ShouldNotBeNil)
	})

	Convey("But new registrations should be allowed", t, func() {
		So(gorgonzola.RegisterDriver("mock2", new(gorgonzola.MockDriver)), ShouldBeNil)
	})

	Convey("But only if the driver is not nil", t, func() {
		So(gorgonzola.RegisterDriver("mock3", nil), ShouldNotBeNil)
	})
}

func TestList(t *testing.T) {
	Convey("List should not return nil", t, func() {
		So(gorgonzola.List(), ShouldNotBeNil)
	})

	Convey("List should contain at least two elements", t, func() {
		So(len(gorgonzola.List()), ShouldBeGreaterThanOrEqualTo, 2)
	})
}
