package regionStats

import (
	"testing"
	"time"

	//"github.com/maxnordlund/breamio/analysis"
	"github.com/maxnordlund/breamio/briee"
	gr "github.com/maxnordlund/breamio/gorgonzola"
)

func TestStatisticsGeneration(t *testing.T) {
	ee := briee.New()

	rs := New(ee, 3*time.Second, 1)

	err := rs.AddRegions(RegionDefinitionMap{
		"middle": RegionDefinition{
			Type:  "square",
			X:     0,
			Y:     0,
			Width: 0.2,
		},
		"error": RegionDefinition{
			Type:  "error",
		},
	})

	if err == nil {
		t.Fatal("AddRegions with an invalid region definition should" +
			"terminate and return an error")
	}

	for i := 0; i < 3; i++ {
		ee.Dispatch("gorgonzola:gazedata", &gr.ETData{
			Filtered: Point2D{
				x: 0.1,
				y: 0.1,
			},
			Timestamp: time.Now(),
		})
	}

	m := rs.generate()

	if m["middle"].Looks != 1 {
		t.Fatal("RegionStatistics should have detected a region look.")
	}
}
