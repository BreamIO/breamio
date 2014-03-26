package regionStats

type AspectMap map[string]RegionDefinitionMap

type RegionDefinitionMap map[string]RegionDefinition

type RegionDefinition struct {
	Type   string  `json:type`
	X      float64 `json:x`
	Width  float64 `json:width`
	Y      float64 `json:y`
	Height float64 `json:"height"`
}
