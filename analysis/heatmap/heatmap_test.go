package heatmap

import (
	"math"
	"image"
	"log"
	"image/color"
	"testing"
	"time"
	//"github.com/maxnordlund/breamio/analysis"
	"github.com/maxnordlund/breamio/briee"
	//gr "github.com/maxnordlund/breamio/gorgonzola"
	"github.com/maxnordlund/breamio/gorgonzola/mock"
	//been "github.com/maxnordlund/breamio/beenleigh"
)

func TestHeatmap(t *testing.T){
	ee := briee.New()
	defer ee.Close()

	subs := ee.Subscribe("heatmap:image", new(image.RGBA)).(<-chan *image.RGBA)

	tracker := mock.New(func(q float64)(x,y float64){
	return 0.5 + 0.5*math.Cos(q), 0.5 + 0.5*math.Sin(q)
	})
	tracker.Connect()
	tracker.Link(ee)
	defer tracker.Close()

	conf := &Config{
		Emitter: 0,
		Duration: time.Minute * 5,
		Hertz: uint(30),
		Res: &Resolution{
			Width: 1920,
			Height: 1080,
			},
		Color: &color.RGBA{
			R: 255,
			G: 0,
			B: 0,
			A: 128,
		},
		}

	//hmr := &HeatmapRun{make(chan struct{})}
	New(ee, conf)
	//heatmap := <-subs
	_ = <-subs
	log.Printf("Done");
	//hmr.Run(tracker)

}
