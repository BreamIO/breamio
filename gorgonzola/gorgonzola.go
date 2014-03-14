package gorgonzola

import (
	"strings"
	"errors"
)

func CreateFromURI(uri string) (Tracker, error) {
	split := strings.SplitN(uri, "://", 2)
	if len(split) < 2 {
		return nil, errors.New("Malformed URI.")
	}
	typ, id := split[0], split[1]
	driver := GetDriver(typ)
	if driver == nil {
		return nil, errors.New("No such driver.")
	}
	return GetDriver(typ).CreateFromId(id)
}

type Point2D interface {
	X() float64
	Y() float64
}

type point2D struct {
	x, y float64
}

func (p point2D) X() float64 {
	return p.x
}

func (p point2D) Y() float64 {
	return p.y
}

func filter(left, right Point2D) Point2D {
	return point2D{(left.X() + right.X()) / 2, (left.Y() + right.Y()) / 2}
}
