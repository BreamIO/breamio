package webber

import (
	"html/template"
)

type Calibrate struct {
	Id          int
	EyeTrackers map[int]string
	Normalize   template.CSS
}

const (
	calibrate = "calibrate.html"
	normalize = "normalize.min.css"
)
