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

type XYer interface {
	X() float64
	Y() float64
}

type Point2D struct {
	Xf, Yf float64
}

func (p Point2D) X() float64 {
	return p.Xf
}

func (p Point2D) Y() float64 {
	return p.Yf
}

func Filter(left, right XYer) XYer {
	return Point2D{(left.X() + right.X()) / 2, (left.Y() + right.Y()) / 2}
}

type Error struct {
	err string
}

func NewError(description string) Error {
	return Error{description}
}

func (err Error) Error() string {
	return err.err
}
