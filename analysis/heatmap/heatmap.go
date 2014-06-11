package heatmap

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"time"

	"github.com/maxnordlund/breamio/analysis"
	been "github.com/maxnordlund/breamio/beenleigh"
	"github.com/maxnordlund/breamio/briee"
	gr "github.com/maxnordlund/breamio/gorgonzola"
)

func init() {
	been.Register(new(HeatmapRun))
}

type Config struct {
	Emitter  int
	Duration time.Duration
	Hertz    uint
	Res      *Resolution
	Color    *color.RGBA
}

type Resolution struct {
	Width  int
	Height int
}

type HeatmapRun struct {
	closeChan chan struct{}
}

func (h *HeatmapRun) Run(logic been.Logic) {
	ee := logic.RootEmitter()

	newHM := ee.Subscribe("new:heatmap", new(Config)).(<-chan *Config)
	defer ee.Unsubscribe("new:heatmap", newHM)

	for {
		select {
		case config := <-newHM:
			NewGenerator(logic.CreateEmitter(config.Emitter), config)
		case <-h.closeChan:
			break
		}
	}
}

func (h *HeatmapRun) Close() error {
	close(h.closeChan)
	return nil
}

const (
	powVar      = 0.5
	limitRadius = 10
)

type Generator struct {
	coordinateHandler analysis.CoordinateHandler
	width, height     int
	publish           chan<- *image.RGBA
	color             *color.RGBA
	closeChan         chan struct{}
}

func NewGenerator(ee briee.EventEmitter, c *Config) *Generator {
	ch := ee.Subscribe("tracker:etdata", &gr.ETData{}).(<-chan *gr.ETData)
	updateSettings := ee.Subscribe("heatmap:update", new(Config)).(<-chan *Config)

	g := &Generator{
		coordinateHandler: analysis.NewCoordBuffer(ch, c.Duration, uint(c.Hertz)),
		width:             c.Res.Width,
		height:            c.Res.Height,
		publish:           ee.Publish("heatmap:image", new(image.RGBA)).(chan<- *image.RGBA),
	}

	if (c.Color == nil) {
		g.color = &color.RGBA{
			R: 255,
			G: 0,
			B: 0,
			A: 128,
		}
	} else {
		g.color = c.Color
	}

	go func() {
		defer ee.Unsubscribe("heatmap:update", updateSettings)
		for {
			select {
			//TODO make sure goroutine can end
			case conf := <-updateSettings:
				g.updateSettings(conf)
			case <-time.After(10 * time.Second):
				g.generate()
			case <-g.closeChan:
				return
			}
		}
	}()

	return g
}

func (gen *Generator) generate() {
	width, height := gen.width, gen.height

	heat := make([][]float64, height)

	for i := range heat {
		heat[i] = make([]float64, width)
	}

	coords := gen.coordinateHandler.GetCoords()

	var maxHeat float64 = 0
	var px, py int
	var x, y float64
	var dist float64

	limSq := float64(limitRadius * limitRadius)

	for coord := range coords {
		f := coord.Filtered
		if valid(f) {
			x = f.X() * float64(width)
			y = f.Y() * float64(height)

			for dx := -limitRadius; dx <= limitRadius; dx++ {
				px = dx + int(x)
				if px >= width || px < 0 {
					continue
				}

				for dy := -limitRadius; dy <= limitRadius; dy++ {
					py = dy + int(y)
					if py >= height || py < 0 {
						continue
					}

					dist = float64(dx*dx + dy*dy)

					if dist <= limSq {
						heat[py][px] += math.Cos(dist / limSq)
						//heat[py][px] += 1
					}
				}
			}
		}
	}

	// Calc max heat to normalize the heat across the map
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
			v := heat[y][x] / maxHeat

			alpha := uint8(float64(gen.color.A) * v)
			heatmap.SetRGBA(x, y, color.RGBA{
				R: gen.color.R,
				G: gen.color.G,
				B: gen.color.B,
				A: uint8(alpha),
				// I don't know what this previous row did, because it overflowed uint8
				//A: uint8(gen.color.A - uint8(float64(gen.color.A)*math.Cos(v*math.Pi))),
			})
		}
	}
	/*
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			v := heat[y][x] / maxHeat

			alpha := uint8(float64(gen.color.A) * v)
			heatmap.SetRGBA(x, y, color.RGBA{
				R: gen.color.R,
				G: gen.color.G,
				B: gen.color.B,
				A: uint8(alpha),
				// I don't know what this previous row did, because it overflowed uint8
				//A: uint8(gen.color.A - uint8(float64(gen.color.A)*math.Cos(v*math.Pi))),
			})
		}
	}*/

	gen.publish <- heatmap
}

/*func colorFor(val float64) (r, g, b byte) {
	return hsl2rgb((1-val)*100, 100, val*50)
}

func hsl2rgb(h, s, l float64) (r, g, b byte) {
	var q, p float64

	if s == 0 {
		ret := byte(l * 255)
		return ret, ret, ret
	}

	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p = 2*l - q

	return byte(hue2rgb(p, q, h+1/3) * 255), byte(hue2rgb(p, q, h) * 255), byte(hue2rgb(p, q, h-1/3) * 255)

}

func hue2rgb(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	} else if t > 1 {
		t -= 1
	}

	if t < 1/6 {
		return p + (q-p)*6*t
	}

	if t < 1/2 {
		return q
	}

	if t < 2/3 {
		return p + (q-p)*(2/3-t)*6
	}

	return p
}*/

func valid(coord gr.XYer) bool {
	return coord.X() < 1 && coord.X() >= 0 && coord.Y() < 1 && coord.Y() >= 0
}

/*
func (gen *Generator) GetCoordinateHandler() *analysis.CoordinateHandler {
	return &gen.coordinateHandler
}
*/

func (gen *Generator) updateSettings(conf *Config) {
	//Config:
	//Emitter int
	//Duration time.Duration
	//Hertz uint
	//Res Resolution
	//Color image.RGBA

	if conf == nil {
		return
	}

	//Don't bother to update Emitter
	if conf.Duration > 0 {
		gen.setDuration(conf.Duration)
	}
	if conf.Hertz > uint(0) {
		gen.setDesiredFreq(conf.Hertz)
	}
	if conf.Res != nil {
		gen.setResolution(conf.Res)
	}
	if conf.Color != nil {
		gen.setColor(conf.Color)
	}
}

func (gen *Generator) setResolution(res *Resolution) {
	if res == nil {
		fmt.Println("Got nil Resolution package")
		return
	}

	gen.height = res.Height
	gen.width = res.Width
}

func (gen *Generator) setDesiredFreq(desiredFreq uint) {
	gen.coordinateHandler.SetDesiredFreq(desiredFreq)
}

func (gen *Generator) setDuration(duration time.Duration) {
	gen.coordinateHandler.SetInterval(duration)
}

func (gen *Generator) setColor(color *color.RGBA) {
	gen.color = color
}
