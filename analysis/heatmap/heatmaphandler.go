package heatmap

import (
	"image/color"
	"time"
)

//HeatMapHandler is an interface for heat map producing modules that are compatible with eriver
type HeatMapHandler interface {
	//The Generate function generates a heatmap which is outputted on the channl given in the constructor.
	//	height and width is the desired dimensions of the heatmap produced.
	Generate(height, width int)
	SetColor(color color.RGBA)
	SetResolution(width, height int)
	SetDesiredFreq(desiredFreq int)
	SetDuration(duration time.Duration)
}
