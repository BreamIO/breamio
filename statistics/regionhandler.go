package main

import (
	"time"
)

//RegionHandler is an interface for modules genarating stats based on regions compatible with eriver

type RegionHandler interface {
	//	Add regions adds all region in a given RegiondefinitioMap to the regions handled by the RegionHandler
	AddRegions(rdm RegionDefinitionMap)

	//Generate generates a RegionStatsMap for all regions in RegionHandler and outputs them on the channel given in the constructor
	Generate()
}

func NewRegionHandler(ee /**EventEmitter*/ int, etID string, duration time.Duration, hertz int) *RegionStatistics {
	return NewRegionStatistics(ee, etID, duration, hertz)
}
