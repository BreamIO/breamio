package regionStats

import (
	"encoding/json"
	"testing"
	"time"

	been "github.com/maxnordlund/breamio/beenleigh"
	"github.com/maxnordlund/breamio/briee"
	gr "github.com/maxnordlund/breamio/gorgonzola"
)

func TestStatisticsGeneration(t *testing.T) {
	ee := briee.New()
	defer ee.Close()

	rs := New(ee, 5*time.Second, 1)

	err := rs.AddRegions(RegionDefinitionMap{
		"middle": RegionDefinition{
			Type:  "square",
			X:     0,
			Y:     0,
			Width: 0.2,
		},
		"error": RegionDefinition{
			Type: "error",
		},
		"afterFailure": RegionDefinition{
			Type:  "square",
			X:     0,
			Y:     0,
			Width: 0.2,
		},
	})

	if err == nil {
		t.Fatal("AddRegions with an invalid region definition should" +
			"terminate and return an error")
	}

	err = rs.AddRegions(RegionDefinitionMap{})

	if err != nil {
		t.Fatal("AddRegions with an empty map should be a problem")
	}

	startTime := time.Now()

	for i := 0; i < 3; i++ {
		ee.Dispatch("tracker:etdata", &gr.ETData{
			Filtered:  gr.Point2D{0.1, 0.1},
			Timestamp: startTime.Add(time.Duration(i*100) * time.Millisecond),
		})
	}

	ee.Dispatch("tracker:etdata", &gr.ETData{
		Filtered:  gr.Point2D{1, 1},
		Timestamp: startTime.Add(1 * time.Second),
	})

	subCh := ee.Subscribe("regionStats:regions", make(RegionStatsMap)).(<-chan RegionStatsMap)

	go rs.Generate()

	m := <-subCh

	if _, ok := m["error"]; ok {
		t.Fatal("The region 'error' should never be created.")
	}

	if _, ok := m["afterFailure"]; ok {
		t.Fatal("The region 'afterFailure' should never be created.")
	}

	if m["middle"].Looks != 1 {
		t.Fatal("RegionStatistics should have detected 1 region look.")
	}

	if time.Duration(m["middle"].TimeInside) != time.Second {
		t.Fatal("The center region was look at for 1 second.")
	}

	bytes, err := json.Marshal(m)

	if err != nil {
		t.Fatal("No error should occur when Marshaling JSON.")
	}

	if string(bytes) != `{"middle":{"looks":1,"time":"00:01"}}` {
		t.Fatal("Marshaling to JSON failed")
	}
}

func TestTimeToString(t *testing.T) {
	if timeToString(12) != "12" {
		t.Fatal("The float 12 toString should be '12'")
	}

	if timeToString(7) != "07" {
		t.Fatal("The float 7 toString should be '07'")
	}
}

func TestWithBeenleigh(t *testing.T) {
	bl := been.New(briee.New)

	
}