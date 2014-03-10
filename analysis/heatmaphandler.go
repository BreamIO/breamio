package statistics

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

//mapOutput is the channel where the heatmap should be delivered when generated
//duration is the time interval that the heatmap should cover
//desiredFreq is the upperbound of the frequency that heatmaphandler will accept
//TODO add some way to tell it where it should listen
func NewHeatMapHandler(ee /**EventEmitter*/ int, etID string, duration time.Duration, desiredFreq, resX, resY int) HeatMapHandler {
	return NewHeatmap(ee, etID, duration, desiredFreq, resX, resY)
}
