package statistics

import (
	"strconv"
	"time"

	"github.com/maxnordlund/breamio/briee"
	// "github.com/maxnordlund/breamio/gorgonzola"
)

type RegionStatistics struct {
	coordinateHandler CoordinateHandler
	regions           []Region
	publish           chan<- RegionStatsMap
}

func NewRegionStatistics(ee briee.EventEmitter, duration time.Duration, hertz int) *RegionStatistics {
	ch := ee.Subscribe("gorgonzola:gazedata", &ETData{}).(<-chan *ETData)

	return &RegionStatistics{
		coordinateHandler: NewCoordinateHandler(ch, duration, hertz),
		regions:           make([]Region, 0),
		publish:           ee.Publish("statistics:regions", make(RegionStatsMap)).(chan<- RegionStatsMap),
	}
}

func (rs RegionStatistics) getCoords() (coords chan *ETData) {
	return rs.coordinateHandler.GetCoords()
}

func (rs *RegionStatistics) AddRegion(name string, def RegionDefinition) {
	rs.regions = append(rs.regions, newRegion(name, def))
}

func (rs *RegionStatistics) AddRegions(defs RegionDefinitionMap) {
	for name, def := range defs {
		rs.regions = append(rs.regions, newRegion(name, def))
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

	//TODO goroutine here somwhere?
	for coord := range rs.getCoords() {
		for i, r := range rs.regions {
			if currentEnterTime[i] == nil && r.Contains(&coord.Filtered) {
				stats[i].Looks++
				currentEnterTime[i] = &coord.Timestamp
			} else if currentEnterTime[i] != nil && !r.Contains(&coord.Filtered) {
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
