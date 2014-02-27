package main

import (
	"math"
)

type Circle struct {
	cx, cy, radius float64
	name           string
}

func NewCircle(name string, cx, cy, radius float64) *Circle {
	return &Circle{
		cx:     cx,
		cy:     cy,
		radius: radius,
		name:   name,
	}
}

func (c Circle) Contains(coord *Coordinate) bool {
	return math.Pow(coord.x-c.cx, 2)+math.Pow(coord.y-c.cy, 2) <= c.radius
}

func (c Circle) RegionName() string {
	return c.name
}

func (c Circle) SetRegionName(name string) {
	c.name = name
}
