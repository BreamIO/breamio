package gorgonzola

import (
	"fmt"
	"io"
	"time"
)

var drivers = make(map[string]Driver)

type Driver interface {
	Create() (Tracker, error)
	CreateS(identifier string) (Tracker, error)
	List() []string
}

type Tracker interface {
	Stream() (<-chan *ETData, <-chan error)
	Connect() error
	io.Closer
	Calibrate(<-chan Point2D, chan<- error)
}

type ETData struct {
	filtered  Point2D
	timestamp time.Time
}

type Point2D struct {
	X, Y float64
}

func List() []string {
	res := make([]string, 0, 32)
	for _, driver := range drivers {
		res = append(res, driver.List()...)
	}
	return res
}

type NotImplementedError string

func (e NotImplementedError) Error() string {
	return fmt.Sprintf("%s is not implemented.", e)
}
