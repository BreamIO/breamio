package gorgonzola_test

import (
	"fmt"
	"github.com/maxnordlund/breamio/gorgonzola"
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
