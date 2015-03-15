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

// The config for the heatmap
type Config struct {
	Emitter  int
	Duration time.Duration //Timespan that the heatmap collects data for
	Hertz    uint          //The frequency that the data should arrive with
	Res      *Resolution   //Resolution of the produced heatmap
	Color    *color.NRGBA  //The color of the heatmap
}

type Resolution struct {
	Width  int
	Height int
}

type HeatmapRun struct {
	closeChan chan struct{}
}

//Start listening for commands of new:heatmap.
//When such arrives, it starts a new heatmap generator
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

//Generator generates heatmaps
type Generator struct {
	coordinateHandler analysis.CoordinateHandler
	width, height     int
	publish           chan<- *image.NRGBA
	color             *color.NRGBA
	closeChan         chan struct{}
}

func NewGenerator(ee briee.EventEmitter, c *Config) *Generator {
	ch := ee.Subscribe("tracker:etdata", &gr.ETData{}).(<-chan *gr.ETData)
	updateSettings := ee.Subscribe("heatmap:update", new(Config)).(<-chan *Config)

	g := &Generator{
		coordinateHandler: analysis.NewCoordBuffer(ch, c.Duration, uint(c.Hertz)),
		width:             c.Res.Width,
		height:            c.Res.Height,
		publish:           ee.Publish("heatmap:image", new(image.NRGBA)).(chan<- *image.NRGBA),
	}

	if c.Color == nil {
		g.color = &color.NRGBA{
			R: 255,
			G: 0,
			B: 0,
			A: 128,
		}
	} else {
		g.color = c.Color
	}

	// Set up a process that listens after configuration updates
	// for this generator.
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

// Make a generator start producing heatmaps.
func (gen *Generator) generate() {
	width, height := gen.width, gen.height

	// heat is how "warm" all the pixels on the screen are
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

	// Go through all coordinates that are in the buffer
	// and increase the heat on the corresponding positions
	// in the "heat" matrix.
	for coord := range coords {
		f := coord.Filtered
		if valid(f) { //f is on screen
			//Calculate the position in the "heat" matrix since
			//the coordinates are normalized [0,1]
			x = f.X() * float64(width)
			y = f.Y() * float64(height)

			//Also increase the heat of all points in a circle around
			//the position. Hence the two for-loops
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
						//A point closer to the center is warmer
						heat[py][px] += math.Cos(dist / limSq)
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

	heatmap := image.NewNRGBA(image.Rect(0, 0, width, height))

	//Draw the heatmap
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			v := heat[y][x] / maxHeat

			alpha := uint8(float64(gen.color.A) * v)
			heatmap.SetNRGBA(x, y, color.NRGBA{
				R: gen.color.R,
				G: gen.color.G,
				B: gen.color.B,
				A: alpha,
			})
		}
	}

	gen.publish <- heatmap
}

func valid(coord gr.XYer) bool {
	return coord.X() < 1 &&
		coord.X() >= 0 &&
		coord.Y() < 1 &&
		coord.Y() >= 0
}

func (gen *Generator) updateSettings(conf *Config) {
	//Config:
	//Emitter int
	//Duration time.Duration
	//Hertz uint
	//Res Resolution
	//Color image.NRGBA

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

func (gen *Generator) setColor(color *color.NRGBA) {
	gen.color = color
}
