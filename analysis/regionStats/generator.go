package statistics

import (
	"strconv"
	"time"

	"github.com/maxnordlund/breamio/analysis"
	"github.com/maxnordlund/breamio/briee"
	gr "github.com/maxnordlund/breamio/gorgonzola"
)

type Generator interface {
	// Adds all regions in a given
	// RegionDefinitionMap for which
	// statistics are to be generated.
	AddRegions(rdm RegionDefinitionMap)

	//Generate generates a RegionStatsMap
	// for all regions registered in the generator
	// and outputs them on a channel
	Generate()
}

type RegionStatistics struct {
	coordinateHandler analysis.CoordinateHandler
	regions           []Region
	publish           chan<- RegionStatsMap
}

func New(ee briee.EventEmitter, duration time.Duration, hertz int) *RegionStatistics {
	ch := ee.Subscribe("gorgonzola:gazedata", &gr.ETData{}).(<-chan *gr.ETData)

	return &RegionStatistics{
		coordinateHandler: analysis.NewCoordBuffer(ch, duration, hertz),
		regions:           make([]Region, 0),
		publish:           ee.Publish("regionStats:regions", make(RegionStatsMap)).(chan<- RegionStatsMap),
	}
}

func (rs RegionStatistics) getCoords() (coords chan *gr.ETData) {
	return rs.coordinateHandler.GetCoords()
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
		region, err := newRegion(name, def)

		if err != nil {
			return err
		}

		rs.regions = append(rs.regions, region)
	}

	return nil
}

// Generates a RegionStatsMap and
// sends it away on the publish channel.
func (rs RegionStatistics) Generate() {
	rs.publish <- rs.generate()
}

func (rs RegionStatistics) generate() RegionStatsMap {
	stats := make([]RegionStatInfo, len(rs.regions))
	currentEnterTime := make([]*time.Time, len(stats))

	//TODO goroutine here somwhere?
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
