package regionStats

import (
	"errors"

	gr "github.com/maxnordlund/breamio/gorgonzola"
)

// Region is an interface describing Regions
type Region interface {
	Contains(gr.XYer) bool
	Name() string
	Update(RegionUpdatePackage)
}

type nameHolder struct {
	name string
}

func (n nameHolder) Name() string {
	return n.name
}

func (n *nameHolder) setName(nam *string) {
	if nam != nil {
		n.name = *nam
	}
}

type point struct {
	x, y float64
}

func (p *point) setX(x *float64) {
	if x != nil {
		p.x = *x
	}
}

func (p *point) setY(y *float64) {
	if y != nil {
		p.y = *y
	}
}

type area struct {
	width, height float64
}

func (a *area) setWidth(w *float64) {
	if w != nil {
		a.width = *w
	}
}

func (a *area) setHeight(h *float64) {
	if h != nil {
		a.height = *h
	}
}

// Create a new Region.
// Error is not nil if unsuccessful.
func newRegion(name string, rd RegionDefinition) (Region, error) {
	switch rd.Type {
	case "square":
		return newSquare(name, rd.X, rd.Y, rd.Width), nil
	case "rectangle":
		return newRectangle(name, rd.X, rd.Y, rd.Width, rd.Height), nil
	case "ellipse":
		return newEllipse(name, rd.X, rd.Y, rd.Width, rd.Height), nil
	case "circle":
		return newCircle(name, rd.X, rd.Y, rd.Width), nil
	}
	return nil, errors.New(rd.Type + " is not a recognized region type.")
}

//Ellipse is a type of region.
type Ellipse struct {
	nameHolder
	point
	area
}

func newEllipse(name string, cx, cy, width, height float64) *Ellipse {
	// Set small width and height values to a small number
	// to avoid devision by zero.
	if width <= 0 {
		width = 0.002
	}

	if height <= 0 {
		height = 0.002
	}

	return &Ellipse{
		nameHolder{name: name},
		point{cx, cy},
		area{width / 2, height / 2},
	}
}

//Loads configuration to the ellipse
func (e *Ellipse) Update(pack RegionUpdatePackage) {
	e.setName(pack.NewName)
	e.setX(pack.X)
	e.setY(pack.Y)
	e.setWidth(pack.Width)
	e.setHeight(pack.Height)
}

func (e *Ellipse) setWidth(w *float64) {
	if w != nil {
		e.area.width = *w / 2
	}
}

func (e *Ellipse) setHeight(h *float64) {
	if h != nil {
		e.area.height = *h / 2
	}
}

func (e Ellipse) Contains(coord gr.XYer) bool {
	dx := coord.X() - e.point.x
	dy := coord.Y() - e.point.y
	w := e.area.width
	h := e.area.height
	return ((dx*dx)/(w*w) + (dy*dy)/(h*h)) < 1
}

// Alias for an Ellipse with the same width and height
func newCircle(name string, cx, cy, radius float64) *Ellipse {
	return newEllipse(name, cx, cy, radius, radius)
}

//Rectangle is a type of region
type Rectangle struct {
	nameHolder
	point
	area
}

func newRectangle(name string, x, y, width, height float64) *Rectangle {
	return &Rectangle{
		nameHolder{name: name},
		point{x, y},
		area{width, height},
	}
}

//Loads configuration to the rectangle
func (e *Rectangle) Update(pack RegionUpdatePackage) {
	e.setName(pack.NewName)
	e.setX(pack.X)
	e.setY(pack.Y)
	e.setWidth(pack.Width)
	e.setHeight(pack.Height)
}

func (r Rectangle) Contains(coord gr.XYer) bool {
	return r.point.x < coord.X() &&
		coord.X() < (r.point.x+r.area.width) &&
		r.point.y < coord.Y() &&
		coord.Y() < (r.point.y+r.area.height)
}

// Alias for a rectangle with equal width and height
func newSquare(name string, x, y, width float64) *Rectangle {
	return newRectangle(name, x, y, width, width)
}
