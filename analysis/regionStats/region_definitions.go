package regionStats

type AspectMap map[string]RegionDefinitionMap

type RegionDefinitionMap map[string]RegionDefinition

type RegionDefinition struct {
	// "square", "rectangle", "circle" or "ellipse"
	Type  string  `json:type`
	X     float64 `json:x`
	Width float64 `json:width`
	Y     float64 `json:y`
	// optional for "square" and "circle" types
	Height float64 `json:height`
}

type RegionDefinitionPackage struct {
	Name string           `json:name`
	Def  RegionDefinition `json:def`
}

type RegionUpdatePackage struct {
	Name    string   `json:name`
	NewName *string  `json:newName`
	X       *float64 `json:x`
	Y       *float64 `json:y`
	Width   *float64 `json:width`
	Height  *float64 `json:height`
}

// type RegionRemovePackage []string
//              simply send ~~^
