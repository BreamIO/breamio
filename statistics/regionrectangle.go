package main

import (
)

type Rectangle struct {
	top, right, bottom, left float64
	name                     string
}

func newRectangle(name string, top, bottom, left, right float64) Rectangle {
	return Rectangle{
		top:    top,
		bottom: bottom,
		left:   left,
		right:  right,
		name:   name,
	}
}

func (r Rectangle) Contains(coord *Coordinate) bool {
	return r.left < coord.x && coord.x < r.right &&
		r.top < coord.y && coord.y < r.bottom
}

func (r Rectangle) RegionName() string {
	return r.name
}

func (r Rectangle) SetRegionName(name string) {
	r.name = name
}

