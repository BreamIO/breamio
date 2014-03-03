package statistics

import (
// 	"github.com/maxnordlund/breamio/gorgonzola"
)

// Region is an interface describing Regions
type Region interface {
	Contains(*Point2D) bool
	Name() string
	SetName(name string)
}

// Create a new Region.
func newRegion(name string, rd RegionDefinition) Region {
	switch rd.Type {
	case "square":
		return newSquare(name, rd.X, rd.Y, rd.Width)
	case "rect":
		return newRectangle(name, rd.X, rd.Y, rd.Width, rd.Height)
	case "ellipse":
		return newEllipse(name, rd.X, rd.Y, rd.Width, rd.Height)
	case "circle":
		return newCircle(name, rd.X, rd.Y, rd.Width)
	default:
		panic("rd.type is unknown: " + rd.Type)
	}
	return nil
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

func (e Ellipse) Contains(coord *Point2D) bool {
	dx := coord.X - e.cx
	dy := coord.Y - e.cy
	return ((dx*dx)/(e.width*e.width) + (dy*dy)/(e.height*e.height)) < 1
}

func (e Ellipse) Name() string {
	return e.name
}

func (e *Ellipse) SetName(name string) {
	e.name = name
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

func (r Rectangle) Contains(coord *Point2D) bool {
	return r.left < coord.X &&
		coord.X < r.right &&
		r.top < coord.Y &&
		coord.Y < r.bottom
}

func (r Rectangle) Name() string {
	return r.name
}

func (r *Rectangle) SetName(name string) {
	r.name = name
}

// Alias for a rectangle with equal width and height
func newSquare(name string, x, y, width float64) *Rectangle {
	return newRectangle(name, x, y, width, width)
}
