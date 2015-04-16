package heatmap

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"
	//"github.com/maxnordlund/breamio/analysis"
	"github.com/maxnordlund/breamio/briee"
	//gr "github.com/maxnordlund/breamio/eyetracker"
	"github.com/maxnordlund/breamio/eyetracker/mock"
	//been "github.com/maxnordlund/breamio/moduler"
)

func TestHeatmap(t *testing.T) {
	ee := briee.New()
	defer ee.Close()

	subs := ee.Subscribe("heatmap:image", new(image.NRGBA)).(<-chan *image.NRGBA)

	tracker := mock.New(func(q float64) (x, y float64) {
		return rand.Float64(), rand.Float64()
		//return 0.5 + 0.5*math.Cos(q), 0.5 + 0.5*math.Sin(q)
	})
	tracker.Connect()
	tracker.Link(ee)
	defer tracker.Close()

	conf := &Config{
		Emitter:  0,
		Duration: time.Minute * 5,
		Hertz:    uint(40),
		Res: &Resolution{
			Width:  600,
			Height: 600,
		},
		Color: &color.NRGBA{
			R: 255,
			G: 0,
			B: 0,
			A: 200,
		},
	}

	//hmr := &HeatmapRun{make(chan struct{})}
	NewGenerator(ee, conf)
	//heatmap := <-subs
	//time.Sleep(time.Millisecond * 60000)
	_ = <-subs
	heatmapImg := <-subs
	saveHeatmap("newHeatmap.png", heatmapImg)
	log.Printf("Done")
	//hmr.Run(tracker)
}

func saveHeatmap(outFilename string, m *image.NRGBA) {
	//outFilename := "blank.png"
	outFile, err := os.Create(outFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()
	log.Print("Saving image to: ", outFilename)
	png.Encode(outFile, m)
}
