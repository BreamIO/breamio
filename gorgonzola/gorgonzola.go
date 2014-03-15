package gorgonzola

import (
	"errors"
	"strings"
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

type XYPoint struct {
	Xf, Yf float64
}

func (p XYPoint) X() float64 {
	return p.Xf
}

func (p XYPoint) Y() float64 {
	return p.Yf
}

func Filter(left, right Point2D) Point2D {
	return XYPoint{(left.X() + right.X()) / 2, (left.Y() + right.Y()) / 2}
}
