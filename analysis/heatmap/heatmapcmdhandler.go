package statistics

import ()

//HeatMapCMDHandler subscribes to an EventEmitter and parses commands which it executes on its corresponding HeatMapHandler
type HeatMapCMDHandler interface {
}

func NewHeatMapCMDHandler(hm HeatMapHandler, ee int /*EventEmitter*/) HeatMapCMDHandler {
	return nil
}
