package statistics

import (
	"testing"
	"time"

	"github.com/maxnordlund/breamio/analysis"
	gr "github.com/maxnordlund/breamio/gorgonzola"
)

func TestStatisticsGeneration(t *testing.T) {
	ich := make(chan *gr.ETData)
	och := make(chan RegionStatsMap)

	rs := RegionStatistics{
		coordinateHandler: analysis.NewCoordBuffer(ich, 3*time.Second, 1),
		regions:           make([]Region, 0),
		publish:           och,
	}

	rs.AddRegions(RegionDefinitionMap{
		"middle": RegionDefinition{
			Type:  "square",
			X:     0,
			Y:     0,
			Width: 0.2,
		},
	})

	for i := 0; i < 3; i++ {
		ich <- &gr.ETData{
			Filtered: Point2D{
				x: 0.1,
				y: 0.1,
			},
			Timestamp: time.Now(),
		}
	}

	m := rs.generate()

	if m["middle"].Looks != 1 {
		t.Fatal("RegionStatistics should have detected a region look.")
	}
}
