package gorgonzola

import (
	"io"
)

type Tracker interface {
	Stream() <-chan *ETData
	Connect() error
	io.Closer
	Calibrate(chan<-Point2D, <-chan error)
}