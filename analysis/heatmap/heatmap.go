package heatmap

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"time"

	"github.com/maxnordlund/breamio/analysis"
	"github.com/maxnordlund/breamio/briee"
	gr "github.com/maxnordlund/breamio/gorgonzola"
)

const (
	powVar      = 0.5
	limitRadius = 25
)

type Generator struct {
	ee                briee.EventEmitter
	coordinateHandler analysis.CoordinateHandler
	width             int
	height            int
	publish           chan<- image.Image
}

func NewGenerator(ee briee.EventEmitter, duration time.Duration, desiredFreq, resX, resY int) *Generator {
	ch := ee.Subscribe("gorgonzola:gazedata", gr.ETData{}).(<-chan *gr.ETData)

	return &Generator{
		ee:                ee,
		coordinateHandler: analysis.NewCoordBuffer(ch, duration, desiredFreq),
		width:             resX,
		height:            resY,
		publish:           ee.Publish("heatmap:image", new(image.Image)).(chan<- image.Image),
	}
}

func (gen *Generator) Generate(height, width int) {
	heat := make([][]int, height)

	for i := range heat {
		heat[i] = make([]int, width)
	}

	coords := gen.coordinateHandler.GetCoords()

	var maxHeat int = 1
	var x, y, px, py, dist int

	for coord := range coords {
		f := coord.Filtered
		if valid(f) {
			x = int(f.X() * float64(width))
			y = int(f.Y() * float64(height))

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
				G: uint8(math.Max(float64(100-(heat[x][y]/maxHeat)), 0)),
				B: 0,
				A: uint8(255 * heat[x][y] / maxHeat),
			})
		}
	}

	gen.publish <- heatmap
}

func valid(coord gr.XYer) bool {
	return coord.X() < 1 && coord.X() >= 0 && coord.Y() < 1 && coord.Y() >= 0
}

func (gen *Generator) GetCoordinateHandler() *analysis.CoordinateHandler {
	return &gen.coordinateHandler
}

func (gen *Generator) SetResolution(width, height int) {
	gen.height = height
	gen.width = width
}

func (gen *Generator) SetDesiredFreq(desiredFreq int) {
	gen.coordinateHandler.SetDesiredFreq(desiredFreq)
}

func (gen *Generator) SetDuration(duration time.Duration) {
	gen.coordinateHandler.SetInterval(duration)
}

func (gen *Generator) SetColor(color color.RGBA) {
	//TODO
	fmt.Println("SetColor is not implemented yet. sob--- T_T")
}
