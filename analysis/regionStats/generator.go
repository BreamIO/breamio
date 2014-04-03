package regionStats

import (
	"errors"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/maxnordlund/breamio/analysis"
	"github.com/maxnordlund/breamio/beenleigh"
	"github.com/maxnordlund/breamio/briee"
	gr "github.com/maxnordlund/breamio/gorgonzola"
)

type Generator interface {
	// Adds all regions in a given
	// RegionDefinitionMap for which
	// statistics are to be generated.
	AddRegions(rdm RegionDefinitionMap)

	// Generate generates a RegionStatsMap
	// for all regions registered in the generator
	// and outputs them on a channel
	Generate()
}

type Config struct {
	Emitter  int
	Duration time.Duration
	Hertz    int
}

// Register in Logic
func init() {
	beenleigh.Register(new(RegionRun))
}

// RegionRun creates and runs generators,
// and terminates them once closed.
type RegionRun struct {
	generators map[int]*RegionStatistics
	close      chan struct{}
}

func (r *RegionRun) Run(logic beenleigh.Logic) {
	var newChan <-chan *Config
	var ee briee.EventEmitter

	ee = logic.RootEmitter()
	newChan = ee.Subscribe("new:regionStats", new(Config)).(<-chan *Config)
	defer ee.Unsubscribe("new:regionStats", newChan)

	for {
		select {
		case rc := <-newChan:
			r.generators[rc.Emitter] =
				New(logic.CreateEmitter(rc.Emitter), rc.Duration, rc.Hertz)
		case <-r.close:
			break
		}
	}
}

func (r *RegionRun) Close() error {
	close(r.close)

	for _, generator := range r.generators {
		close(generator.close)
	}

	return nil
}

type RegionStatistics struct {
	coordinates *analysis.CoordBuffer
	regions     []Region
	publish     chan<- RegionStatsMap
	close       chan struct{}
}

func New(ee briee.PublishSubscriber, duration time.Duration, hertz int) *RegionStatistics {
	ch := ee.Subscribe("tracker:etdata", &gr.ETData{}).(<-chan *gr.ETData)

	addRegionCh := ee.Subscribe("regionStats:addRegion", new(RegionDefinitionPackage)).(<-chan *RegionDefinitionPackage)
	updateRegionCh := ee.Subscribe("regionStats:updateRegion", new(RegionUpdatePackage)).(<-chan *RegionUpdatePackage)
	removeRegionCh := ee.Subscribe("regionStats:removeRegion", make([]string, 0, 0)).(<-chan []string)

	log := log.New(os.Stderr, "[ RegionStats ]", log.LstdFlags)

	rs := &RegionStatistics{
		coordinates: analysis.NewCoordBuffer(ch, duration, hertz),
		regions:     make([]Region, 0),
		publish:     ee.Publish("regionStats:regions", make(RegionStatsMap)).(chan<- RegionStatsMap),
	}

	go func() {
		for {
			select {
			case <-rs.close:
				close(rs.publish)
				ee.Unsubscribe("regionStats:addRegions", addRegionCh)
				ee.Unsubscribe("regionStats:updateRegions", updateRegionCh)
				ee.Unsubscribe("regionStats:removeRegions", removeRegionCh)
				ee.Unsubscribe("tracker:etdata", ch)
				return

			case regionDef := <-addRegionCh:
				err := rs.AddRegion(regionDef.Name, regionDef.Def)
				if err != nil {
					log.Println(err.Error())
				}

			case regionUpdate := <-updateRegionCh:
				err := rs.UpdateRegion(regionUpdate)
				if err != nil {
					log.Println(err.Error())
				}

			case regs := <-removeRegionCh:
				rs.RemoveRegions(regs)

			default:
				time.Sleep(time.Second)
				rs.Generate()
			}
		}
	}()

	return rs
}

func (rs RegionStatistics) getCoords() (coords chan *gr.ETData) {
	return rs.coordinates.GetCoords()
}

func (rs *RegionStatistics) AddRegion(name string, def RegionDefinition) error {
	region, err := newRegion(name, def)

	if err != nil {
		return err
	}

	rs.regions = append(rs.regions, region)

	return nil
}

func (rs *RegionStatistics) AddRegions(defs RegionDefinitionMap) error {
	for name, def := range defs {
		err := rs.AddRegion(name, def)

		if err != nil {
			return err
		}
	}

	return nil
}

func (r *RegionStatistics) UpdateRegion(pack *RegionUpdatePackage) error {
	if pack == nil {
		return errors.New("Got nil update package.")
	}

	for _, region := range r.regions {
		if region.Name() == pack.Name {
			region.Update(*pack)
			return nil
		}
	}

	return errors.New("No such region: " + pack.Name)
}

func (rs *RegionStatistics) RemoveRegions(regs []string) {
	for _, name := range regs {
		for i, reg := range rs.regions {
			if reg.Name() == name {
				rs.regions = append(
					rs.regions[:i],
					rs.regions[i+1:]...)
				break
			}
		}
	}
}

// Generates a RegionStatsMap and
// sends it away on the publish channel.
func (rs RegionStatistics) Generate() {
	rs.publish <- rs.generate()
}

func (rs RegionStatistics) generate() RegionStatsMap {
	stats := make([]RegionStatInfo, len(rs.regions))
	currentEnterTime := make([]*time.Time, len(stats))

	for coord := range rs.getCoords() {
		for i, r := range rs.regions {
			if currentEnterTime[i] == nil && r.Contains(coord.Filtered) {
				stats[i].Looks++
				currentEnterTime[i] = &coord.Timestamp
			} else if currentEnterTime[i] != nil && !r.Contains(coord.Filtered) {
				stats[i].TimeInside += InsideTime(coord.Timestamp.Sub(*currentEnterTime[i]))
				currentEnterTime = nil
			}
		}
	}

	var retMap = make(RegionStatsMap)

	for i, r := range rs.regions {
		retMap[r.Name()] = stats[i]
	}

	return retMap
}

type RegionStatsMap map[string]RegionStatInfo

//TODO rename
type RegionStatInfo struct {
	Looks      int        `json:"looks"`
	TimeInside InsideTime `json:"time"`
}

type RegionStats map[string]RegionStats

type InsideTime time.Duration

func (it InsideTime) MarshalJSON() ([]byte, error) {
	return []byte("\"" + timeToString(time.Duration(it).Minutes()) +
		":" + timeToString(time.Duration(it).Seconds()) + "\""), nil
}

func timeToString(t float64) string {
	tmp := strconv.Itoa(int(t) % 60)
	if len(tmp) == 1 {
		return "0" + tmp
	}
	return tmp
}
