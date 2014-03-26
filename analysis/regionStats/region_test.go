package statistics

import (
	"testing"
)

type Point2D struct {
	x, y float64
}

func (p Point2D) X() float64 {
	return p.x
}

func (p Point2D) Y() float64 {
	return p.y
}

func TestCircle(t *testing.T) {
	reg, _ := newRegion("center", RegionDefinition{
		Type:  "circle",
		X:     0.5,
		Y:     0.5,
		Width: 0.5,
	})

	if !reg.Contains(&Point2D{0.5, 0.5}) {
		t.Fatal("Circle should contain center of itself!")
	}

	if reg.Contains(&Point2D{0.0, 0.5}) {
		t.Fatal("Circle shouldn't contain edge!")
	}

	if reg.Name() != "center" {
		t.Fatal("Name getter should work!")
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

	if !reg.Contains(&Point2D{0.1, 0.1}) {
		t.Fatal("Ellipse should contain 0.1 : 0.1!")
	}

	if reg.Name() != "upper-left" {
		t.Fatal("Name getter should work!")
	}
}

func TestSquare(t *testing.T) {
	reg, _ := newRegion("center", RegionDefinition{
		Type:  "square",
		X:     0.25,
		Y:     0.25,
		Width: 0.5,
	})

	if !reg.Contains(&Point2D{0.5, 0.5}) {
		t.Fatal("Square should contain center of itself!")
	}

	if !reg.Contains(&Point2D{0.26, 0.26}) {
		t.Fatal("Square should points within itself!")
	}

	if reg.Contains(&Point2D{0.25, 0.25}) {
		t.Fatal("Square shouldn't contain edge!")
	}

	if reg.Name() != "center" {
		t.Fatal("Name getter should work!")
	}
}

func TestRectangle(t *testing.T) {
	reg, _ := newRegion("center", RegionDefinition{
		Type:   "rect",
		X:      0.25,
		Y:      0.25,
		Width:  0.5,
		Height: 0.1,
	})

	if !reg.Contains(&Point2D{0.5, 0.3}) {
		t.Fatal("Rectangle should contain center of itself!")
	}

	if !reg.Contains(&Point2D{0.26, 0.26}) {
		t.Fatal("Rectangle should points within itself!")
	}

	if reg.Contains(&Point2D{0.25, 0.25}) {
		t.Fatal("Rectangle shouldn't contain edge!")
	}

	if reg.Name() != "center" {
		t.Fatal("Name getter should work!")
	}
}
