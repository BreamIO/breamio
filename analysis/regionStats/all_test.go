package regionStats

import (
	"encoding/json"
	been "github.com/maxnordlund/breamio/beenleigh"
	"github.com/maxnordlund/breamio/briee"
	gr "github.com/maxnordlund/breamio/gorgonzola"
	"github.com/maxnordlund/breamio/gorgonzola/mock"
	"log"
	"math"
	"testing"
	"time"
	//"strconv"
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
	re := bl.RootEmitter()

	tracker := mock.New(func(q float64) (x, y float64) {
		return 0.5 + 0.5*math.Cos(q), 0.5 + 0.5*math.Sin(q)
	})

	tracker.Connect()
	tracker.Link(re)
	//defer tracker.Close()

	go bl.ListenAndServe()
	pub := re.Publish("new:regionStats", new(Config)).(chan<- *Config)
	sub := re.Subscribe("regionStats:regions", make(RegionStatsMap)).(<-chan RegionStatsMap)
	// Add new region
	pub <- &Config{
		Emitter:  256,
		Duration: time.Second * 20,
		Hertz:    40,
	}

	var dispatchAddRegion = func(Name string, Type string, X float64, Y float64, Width float64, Height float64) {
		re.Dispatch("regionStats:addRegion", &RegionDefinitionPackage{
			Name: Name,
			Def: RegionDefinition{
				Type:   Type,
				X:      X,
				Width:  Width,
				Y:      Y,
				Height: Height,
			},
		})
	}

	// name, type, X, Y, width, height
	/*
		for i:= 0; i<20; i++ {
			dispatchAddRegion(strconv.Itoa(i),"circle", 0.5, 0.5, 1.0, 1.0)
		}
	*/

	dispatchAddRegion("bottom-right", "rect", 0.5, 0.5, 0.5, 0.5)

	/*
		re.Dispatch("regionStats:addRegion", &RegionDefinitionPackage{
			Name: "upper-left",
			Def: RegionDefinition{
				Type: "square",
				Width: 0.5,
			},
		})

		re.Dispatch("regionStats:addRegion", &RegionDefinitionPackage{
			Name: "all",
			Def: RegionDefinition{
				Type: "square",
				Width: 1.0,
			},
		})
	*/
	timeout := time.After(50000 * time.Millisecond)
	omgquit := false
	for !omgquit {
		select {
		case regiondata := <-sub:
			log.Println(regiondata)
			bytes, _ := json.Marshal(regiondata)
			log.Println(string(bytes))
		case <-timeout:
			omgquit = true
		}
	}

	//re.Dispatch("shutdown", struct{}{});
	log.Println("Done!")

}

/*
func TestWithBeenleigh(t *testing.T) {
	bl := been.New(briee.New)

	_ = bl.RootEmitter()

	ee2 := bl.CreateEmitter(777)

	pub := ee.Publish("new:RegionStats", Config{}).(chan<- Config);
	//pub := ee.Publish("new:RegionStats", Config{	777, time.Second * 5, 1	}).(chan<- Config)

	reg := ee2.Subscribe("regionStats:addRegion", new(RegionDefinitionPackage)).(<-chan *RegionDefinitionPackage)

	// ee2.Dispatch("regionStats:updateRegion", nil)

	ee2.Dispatch("regionStats:addRegion", &RegionDefinitionPackage{
		Name: "upper-left",
		Def: RegionDefinition{
			Type: "square",
			Width: 0.5,
		},
	})

	// startTime := time.Now()

	// ee2.Dispatch("tracker:etdata", &gr.ETData{
	// 	Filtered:  gr.Point2D{0.1, 0.1},
	// 	Timestamp: startTime,
	// })

	// ee2.Dispatch("tracker:etdata", &gr.ETData{
	// 	Filtered:  gr.Point2D{1, 1},
	// 	Timestamp: startTime.Add(time.Second),
	// })

	bytes, err := json.Marshal(<-reg)

	if err != nil {
		t.Fatal("No error should occur when Marshaling JSON.")
	}

	if string(bytes) != `{"upper-left":{"looks":1,"time":"00:01"}}` {
		t.Fatal("Marshaling to JSON failed")
	}
}*/
