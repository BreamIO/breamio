package eyetribe

import (
	"github.com/zephyyrr/thegotribe"
)

type p2d struct {
	thegotribe.Point2D
}

func (p p2d) X() float64 {
	return p.Point2D.X
}

func (p p2d) Y() float64 {
	return p.Point2D.Y
}
