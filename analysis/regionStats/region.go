package regionStats

import (
	"errors"

	gr "github.com/maxnordlund/breamio/gorgonzola"
)

// Region is an interface describing Regions
type Region interface {
	Contains(gr.XYer) bool
	Name() string
}

// Create a new Region.
// Error is not nil if unsuccessful.
func newRegion(name string, rd RegionDefinition) (Region, error) {
	switch rd.Type {
	case "square":
		return newSquare(name, rd.X, rd.Y, rd.Width), nil
	case "rect":
		return newRectangle(name, rd.X, rd.Y, rd.Width, rd.Height), nil
	case "ellipse":
		return newEllipse(name, rd.X, rd.Y, rd.Width, rd.Height), nil
	case "circle":
		return newCircle(name, rd.X, rd.Y, rd.Width), nil
	}
	return nil, errors.New(rd.Type + " is not a recognized region type.")
}

type Ellipse struct {
	cx, cy, width, height float64
	name                  string
}

func newEllipse(name string, cx, cy, width, height float64) *Ellipse {
	// Set small width and height values to a small number
	// to avoid devision by zero.
	if width <= 0 {
		width = 0.001
	}

	if height <= 0 {
		height = 0.001
	}

	return &Ellipse{
		cx:     cx,
		cy:     cy,
		height: height,
		width:  width,
		name:   name,
	}
}

func (e Ellipse) Contains(coord gr.XYer) bool {
	dx := coord.X() - e.cx
	dy := coord.Y() - e.cy
	return ((dx*dx)/(e.width*e.width) + (dy*dy)/(e.height*e.height)) < 1
}

func (e Ellipse) Name() string {
	return e.name
}

// Alias for an Ellipse with the same width and height
func newCircle(name string, cx, cy, radius float64) *Ellipse {
	return newEllipse(name, cx, cy, radius, radius)
}

type Rectangle struct {
	top, right, bottom, left float64
	name                     string
}

func newRectangle(name string, x, y, width, height float64) *Rectangle {
	return &Rectangle{
		left:   x,
		top:    y,
		right:  x + width,
		bottom: y + height,
		name:   name,
	}
}

func (r Rectangle) Contains(coord gr.XYer) bool {
	return r.left < coord.X() &&
		coord.X() < r.right &&
		r.top < coord.Y() &&
		coord.Y() < r.bottom
}

func (r Rectangle) Name() string {
	return r.name
}

// Alias for a rectangle with equal width and height
func newSquare(name string, x, y, width float64) *Rectangle {
	return newRectangle(name, x, y, width, width)
}
