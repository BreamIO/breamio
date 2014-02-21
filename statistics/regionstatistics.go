package main

import (
	"strconv"
	"time"
)

type RegionStatistics struct {
	coordinateHandler CoordinateHandler
	regions           []Region
	publish           chan<- *RegionStatsMap
}

func NewRegionStatistics(ee /**EventEmitter*/ int, etID string, duration time.Duration, hertz int) *RegionStatistics {
	return &RegionStatistics{
		coordinateHandler: NewCoordinateHandler(make(chan *Coordinate) /*ee.Subscribe(etID, Coordinate{}).(<-chan *Coordinate)*/, duration, hertz),
		regions:           make([]Region, 0),
		publish:           make(chan<- *RegionStatsMap), /*ee.Publish(etID + ":stats", *RegionStatsMap)*/
	}
}

func (rs RegionStatistics) getCoords() (coords chan *Coordinate) {
	return rs.coordinateHandler.GetCoords()
}

func (rs *RegionStatistics) AddRegions(rdm RegionDefinitionMap) {
	for name, def := range rdm {
		rs.regions = append(rs.regions, newRegion(name, def))
	}
}

func (rs RegionStatistics) Generate() {
	coords := rs.getCoords()
	var stats = make([]RegionStatInfo, len(rs.regions))
	currentEnterTime := make([]*time.Time, len(stats))

	//TODO goroutine here somwhere?
	for coord := range coords {
		for i, r := range rs.regions {
			if currentEnterTime[i] == nil && r.Contains(coord) {
				stats[i].Looks++
				currentEnterTime[i] = &coord.timestamp
			} else if currentEnterTime[i] != nil && !r.Contains(coord) {
				stats[i].TimeInside += InsideTime(coord.timestamp.Sub(*currentEnterTime[i]))
				currentEnterTime = nil
			}
		}
	}

	var retMap = make(RegionStatsMap)

	for i, r := range rs.regions {
		retMap[r.RegionName()] = stats[i]
	}

	rs.publish <- &retMap
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
	return []byte("\"" + timeToString(time.Duration(it).Minutes()) + ":" + timeToString(time.Duration(it).Seconds()) + "\""), nil
}

func timeToString(t float64) string {
	tmp := strconv.Itoa(int(t) % 60)
	if len(tmp) == 1 {
		return "0" + tmp
	}
	return tmp
}
