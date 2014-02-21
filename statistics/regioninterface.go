package main

/*
Region is an interface describing Regions used by RegionHandler
*/
type Region interface {
	Contains(*Coordinate) bool //Returns true iff the coordinate is inside or on the borders of Region
	RegionName() string        //Returns the name of the region
	SetRegionName(name string) //Sets the name of Region
}

/*
Create a new Region module
name is the name of the region and rd the definitions of the region
*/
func newRegion(name string, rd RegionDefinition) Region {
	switch rd.Type {
	case "rect":
		return newRectangle(name, rd.Y, rd.Y+rd.Height, rd.X, rd.X+rd.Width)
	case "circle":
	default:
		panic("rd.type is unknown: " + rd.Type)
	}
	return nil
}
