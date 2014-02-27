package main

import (

)

//RegionCMDHandler subscribes to an EventEmitter and parses commands which it executes on its corresponding RegionHandler
type RegionCMDHandler interface {

}

func NewRegionCMDHandler(rh RegionHandler, ee int /*EventEmitter*/) RegionCMDHandler {
	return nil
}
