package main

import (
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"time"
)

const (
	powVar      = 0.5
	limitRadius = 25
)

type HeatMap struct {
	coordinateHandler CoordinateHandler
	width int
	height int
}

func NewHeatmap(ee /*EventEmitter*/ int, duration time.Duration, desiredFreq, resX, resY int) *HeatMap {
	return &HeatMap{
		coordinateHandler: NewCoordinateHandler(make(chan Coordinate), duration, desiredFreq),
		width: resX,
		height: resY,
	}
}

func (hm HeatMap) Generate(height, width int) image.Image {
	heat := make([][]int, height)

	for i := range heat {
		heat[i] = make([]int, width)
	}

	coords := (*hm.GetCoordinateHandler()).GetCoords()

	var maxHeat int = 1
	var x, y, px, py, dist int

	for coord := range coords {
		if valid(coord) {
			x = int(coord.x * float64(width))
			y = int(coord.y * float64(height))

			for dx := -limitRadius; dx <= limitRadius; dx++ {
				px = dx + x
				if px >= width || px < 0 {
					continue
				}

				for dy := -limitRadius; dy <= limitRadius; dy++ {
					py = dy + y
					if py >= width || py < 0 {
						continue
					}

					dist = dx*dx + dy*dy + 1

					if math.Sqrt(float64(dist)) <= limitRadius {
						heat[px][py] += int(float64(100) / math.Pow(float64(dist), powVar))

					}
				}
			}
		}
	}


	for x, col := range heat {
		for y := range col {
			if heat[x][y] > maxHeat {
				maxHeat = heat[x][y]
			}
		}
	}

	heatmap := image.NewRGBA(image.Rect(0, 0, width, height))

	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			heatmap.SetRGBA(x, y, color.RGBA{
				R: 200,
				G: uint8(math.Max(float64(100-(heat[x][y]/ maxHeat)), 0)),
				B: 0,
				A: uint8(255 * heat[x][y] / maxHeat),
			})
		}
	}
	png.Encode(os.Stdout, heatmap)	
	return heatmap
}

func valid(coord *Coordinate) bool {
	return coord.x < 1 && coord.x >= 0 && coord.y < 1 && coord.y >= 0
}

func (hm HeatMap) GetCoordinateHandler() *CoordinateHandler {
	return hm.GetCoordinateHandler()
}


func (hm HeatMap) SetResolution(width, height int) {
	hm.height = height
	hm.width = width
}

func (hm HeatMap) SetDesiredFreq(desiredFreq int) {
	hm.coordinateHandler.SetDesiredFreq(desiredFreq)
}

func (hm HeatMap) SetDuration(duration time.Duration) {
	hm.coordinateHandler.SetDuration(duration)
}









