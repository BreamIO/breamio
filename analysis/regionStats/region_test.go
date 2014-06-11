package regionStats

import (
	"testing"

	gr "github.com/maxnordlund/breamio/gorgonzola"
)

func TestCircle(t *testing.T) {
	reg, _ := newRegion("center", RegionDefinition{
		Type:  "circle",
		X:     0.5,
		Y:     0.5,
		Width: 0.5,
	})

	if !reg.Contains(&gr.Point2D{0.5, 0.5}) {
		t.Fatal("Circle should contain center of itself!")
	}

	if reg.Contains(&gr.Point2D{0.0, 0.5}) {
		t.Fatal("Circle shouldn't contain edge!")
	}

	if reg.Name() != "center" {
		t.Fatal("Name getter should work!")
	}

	reg.Update(RegionUpdatePackage{
		Name:    "center",
		NewName: strAddr("middle"),
		X:       f64Addr(0),
	})

	if reg.Name() != "middle" {
		t.Fatal("Update should work!")
	}

	if reg.Contains(&gr.Point2D{0.5, 0.5}) {
		t.Fatal("Update should work!")
	}
}

func TestEllipse(t *testing.T) {
	reg, _ := newRegion("upper-left", RegionDefinition{
		Type:   "ellipse",
		X:      0.0,
		Y:      0.0,
		Width:  0.5,
		Height: 0.25,
	})

	if !reg.Contains(&gr.Point2D{0.1, 0.1}) {
		t.Fatal("Ellipse should contain 0.1 : 0.1!")
	}

	if reg.Name() != "upper-left" {
		t.Fatal("Name getter should work!")
	}

	reg, _ = newRegion("miniscule", RegionDefinition{
		Type: "ellipse",
		X:    0.5,
		Y:    0.5,
	})

	if !reg.Contains(&gr.Point2D{0.5, 0.5}) {
		t.Fatal("Ellipse should contain it's center.")
	}

	if reg.Contains(&gr.Point2D{0.501, 0.501}) {
		t.Fatal("Ellipse shouldn't contain points on the edge.")
	}
}

func TestSquare(t *testing.T) {
	reg, _ := newRegion("center", RegionDefinition{
		Type:  "square",
		X:     0.25,
		Y:     0.25,
		Width: 0.5,
	})

	if !reg.Contains(&gr.Point2D{0.5, 0.5}) {
		t.Fatal("Square should contain center of itself!")
	}

	if !reg.Contains(&gr.Point2D{0.26, 0.26}) {
		t.Fatal("Square should points within itself!")
	}

	if reg.Contains(&gr.Point2D{0.25, 0.25}) {
		t.Fatal("Square shouldn't contain edge!")
	}

	if reg.Name() != "center" {
		t.Fatal("Name getter should work!")
	}
}

func TestRectangle(t *testing.T) {
	reg, _ := newRegion("center", RegionDefinition{
		Type:   "rectangle",
		X:      0.25,
		Y:      0.25,
		Width:  0.5,
		Height: 0.1,
	})

	if !reg.Contains(&gr.Point2D{0.5, 0.3}) {
		t.Fatal("Rectangle should contain center of itself!")
	}

	if !reg.Contains(&gr.Point2D{0.26, 0.26}) {
		t.Fatal("Rectangle should points within itself!")
	}

	if reg.Contains(&gr.Point2D{0.25, 0.25}) {
		t.Fatal("Rectangle shouldn't contain edge!")
	}

	if reg.Name() != "center" {
		t.Fatal("Name getter should work!")
	}
}

func TestErronous(t *testing.T) {
	reg, err := newRegion("error", RegionDefinition{
		Type: "error",
	})

	if err == nil {
		t.Fatal("'error' shouldn't be a valid region type!")
	}

	if reg != nil {
		t.Fatal("newRegion shouldn't return a region if an error occurs.")
	}
}

func strAddr(s string) *string {
	return &s
}

func f64Addr(s float64) *float64 {
	return &s
}
