package regionStats

import (
	"encoding/json"
	"github.com/maxnordlund/breamio/briee"
	gr "github.com/maxnordlund/breamio/eyetracker"
	"github.com/maxnordlund/breamio/eyetracker/mock"
	been "github.com/maxnordlund/breamio/moduler"
	"log"
	"math"
	"testing"
	"time"
	//"strconv"
)

func TestStatisticsGeneration(t *testing.T) {
	ee := briee.New()
	defer ee.Close()

	rs := New(ee, "5s", 1, "1s")

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
		Duration: "20s",
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

	dispatchAddRegion("bottom-right", "rectangle", 0.5, 0.5, 0.5, 0.5)

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
	timeout := time.After(2 * time.Millisecond)
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

func TestBLwithBufferCommands(t *testing.T) {
	bl := been.New(briee.New)
	re := bl.RootEmitter()

	tracker := mock.New(func(q float64) (x, y float64) {
		return 0.5 + 0.5*math.Cos(q), 0.5 + 0.5*math.Sin(q)
	})

	tracker.Connect()
	tracker.Link(re)

	go bl.ListenAndServe()
	pub := re.Publish("new:regionStats", new(Config)).(chan<- *Config)
	sub := re.Subscribe("regionStats:regions", make(RegionStatsMap)).(<-chan RegionStatsMap)
	// Add new region
	pub <- &Config{
		Emitter:  256, // Emitter ID
		Duration: "20s",
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

	dispatchAddRegion("bottom-right", "rectangle", 0.5, 0.5, 0.5, 0.5)

	// The generator should gathering data as default
	regiondata := <-sub
	log.Println(regiondata)

	// Stop the gathering data
	re.Dispatch("regionStats:stop", struct{}{})
	timeout := time.After(1 * time.Second)
	select {
	case regiondata = <-sub:
		log.Printf("Error")
	case <-timeout:
		// This is the success case
		break
	}

	/*broken test
	// Start the gathering again
	re.Dispatch("regionStats:start", struct{}{})
	//pause
	re.Dispatch("regionStats:generate", struct{}{})
	regiondata = <-sub
	bytes, err := json.Marshal(regiondata)
	if string(bytes) != `{"bottom-right":{"looks":0,"time":"00:00"}}` {
		t.Fatal("Does not collect data after start")
	}
	log.Println(regiondata)

	// Restart/Flush
	re.Dispatch("regionStats:restart", struct{}{})
	regiondata = <-sub
	log.Println(regiondata)
	*/

	//re.Dispatch("shutdown", struct{}{});
	log.Println("Done!")
}

func TestInRange(t *testing.T) {
	t.Log("Testing InRange")
	p1 := gr.Point2D{
		Xf: 0.0,
		Yf: 0.0,
	}

	p2 := gr.Point2D{
		Xf: 0.3,
		Yf: 0.4,
	}

	if inRange(p1, p2, 0.5) != true {
		t.Fail()
	}

}

func TestNewFixation(t *testing.T) {
	p1 := gr.Point2D{
		Xf: 0.0,
		Yf: 0.0,
	}

	p2 := gr.Point2D{
		Xf: 0.5,
		Yf: 0.5,
	}

	p3 := newFixation(p1, p2, 2)
	if p3.X() != 0.25 || p3.Y() != 0.25 {
		t.Fail()
	}

	p4 := newFixation(p2, p1, 2)
	if p4.X() != 0.25 || p4.Y() != 0.25 {
		t.Fail()
	}

}

func TestNewFixation3Points(t *testing.T) {
	p1 := gr.Point2D{
		Xf: 0.0,
		Yf: 0.0,
	}

	p2 := gr.Point2D{
		Xf: 1.0,
		Yf: 0.0,
	}

	p3 := gr.Point2D{
		Xf: 0.5,
		Yf: 1.0,
	}

	p4 := newFixation(p1, p2, 2)
	t.Log(p4.X())
	t.Log(p4.Y())
	p5 := newFixation(p4, p3, 3)

	t.Log(p5.X())
	t.Log(p5.Y())

	if p5.X() != 0.5 || p5.Y() != (1.0/3.0) {
		t.Fail()
	}
}
