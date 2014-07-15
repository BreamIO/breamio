package webber

import (
	"html/template"
)

const (
	calibrate = "calibrate.html"
	normalize = "normalize.min.css"
)

type Calibrate struct {
	Id          int
	EyeTrackers map[int]string
	Normalize   template.CSS
}

type drawer struct {
	Id int
}

type statistics struct {
	Id int
}

type consumer struct {
	// TODO: Add fields
}
