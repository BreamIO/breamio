package regionStats

import (
	"errors"
	"github.com/maxnordlund/breamio/comte"
	"github.com/mitchellh/mapstructure"
	"math"
	"strconv"
	"time"

	"github.com/maxnordlund/breamio/analysis"
	"github.com/maxnordlund/breamio/beenleigh"
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
	Emitter            int
	Duration           string //The time given parsable bu time.ParseDuration(string), e.g. 2h34m2s3ms
	Hertz              uint
	GenerationInterval string //Works the same as duration
}

// Register in Logic
func init() {
	beenleigh.Register(&Factory{
		generators: make(map[int]*RegionStatistics),
		closeChan:  make(chan struct{}),
	})
}

// RegionRun creates generators,
// and terminates them when closing.
type Factory struct {
	generators map[int]*RegionStatistics
	closeChan  chan struct{}
}

func (Factory) String() string {
	return "RegionStats"
}

func (r *Factory) New(c beenleigh.Constructor) beenleigh.Module {
	c.Logger.Println("Starting a new generator for emitter:", c.Emitter)

	r.generators[c.Emitter] = New(c)
	return r.generators[c.Emitter]
}

func (r *Factory) Close() error {
	close(r.closeChan)

	for _, generator := range r.generators {
		close(generator.closeChan)
	}

	return nil
}

func (r *Factory) Run(l beenleigh.Logic) {
	confs := make([]Config, 0)
	comte.Section(r.String(), &confs)

	for _, conf := range confs {
		c := beenleigh.Constructor{
			Logic:   l,
			Logger:  beenleigh.NewLogger(r),
			Emitter: conf.Emitter,
			Parameters: map[string]interface{}{
				"Duration":           conf.Duration,
				"GenerationInterval": conf.GenerationInterval,
				"Hertz":              conf.Hertz,
			}, //Not beautiful, but necessary evil.
		} // TODO, add a Logic.NewConstructor(emitter id, params interface{}) Constructor
		c.Logger.Println("Starting a new generator for emitter:", c.Emitter)
		r.generators[c.Emitter] = New(c)
	}

	beenleigh.RunFactory(l, r)
}

type RegionStatistics struct {
	beenleigh.SimpleModule

	coordinates        *analysis.CoordBuffer
	regions            []Region
	dataCh             chan<- *gr.ETData
	closeChan          chan struct{}
	generationInterval time.Duration

	//Event Export Declarations

	MethodOnETData     beenleigh.EventMethod `event:"tracker:etdata"`
	MethodAddRegion    beenleigh.EventMethod `returns:"AddRegion:error"`
	MethodUpdateRegion beenleigh.EventMethod `returns:"UpdateRegion:error"`
	MethodRemoveRegion beenleigh.EventMethod `returns:"RemoveRegion:error"`
	MethodGetRegions   beenleigh.EventMethod `returns:"Regions"`

	MethodStart  beenleigh.EventMethod
	MethodStop   beenleigh.EventMethod
	MetodRestart beenleigh.EventMethod
}

func New(c beenleigh.Constructor) *RegionStatistics {
	var rc Config
	mapstructure.Decode(c.Parameters, &rc)

	duration, err := time.ParseDuration(rc.Duration)
	if err != nil {
		c.Logger.Println(err, "Defaulting duration to 60 seconds")
		duration = time.Minute
	}
	generationInterval, err := time.ParseDuration(rc.GenerationInterval)
	if err != nil {
		c.Logger.Println(err, "Defaulting generation interval to 60 seconds")
		duration = time.Minute
	}

	datach := make(chan *gr.ETData)

	rs := &RegionStatistics{
		SimpleModule:       beenleigh.NewSimpleModule("RegionStats", c),
		coordinates:        analysis.NewCoordBuffer(datach, duration, rc.Hertz),
		regions:            make([]Region, 0),
		dataCh:             datach,
		closeChan:          make(chan struct{}),
		generationInterval: generationInterval,
	}

	go func(generationInterval time.Duration) {
		genTicker := time.Tick(generationInterval)
		emitter := c.Logic.CreateEmitter(c.Emitter)
		publisher := emitter.Publish(rs.String()+":Stats", make(RegionStatsMap)).(chan<- RegionStatsMap)
		defer close(publisher)
		for {
			select {
			case <-genTicker:
				rs.generate(publisher)
			case <-rs.closeChan:
				return
			}
		}
	}(rs.generationInterval)

	return rs
}

func (rs RegionStatistics) OnETData(data *gr.ETData) {
	rs.dataCh <- data
}

func (rs RegionStatistics) getCoords() (coords chan *gr.ETData) {
	return rs.coordinates.GetCoords()
}

func (rs *RegionStatistics) AddRegion(pack *RegionDefinitionPackage) error {
	if pack == nil {
		return errors.New("Got nil RegionDefinitionPackage.")
	}

	region, err := newRegion(pack.Name, pack.Def)

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

func (r *RegionStatistics) UpdateRegion(pack *RegionUpdatePackage) error {
	if pack == nil {
		return errors.New("Got nil RegionUpdatePackage.")
	}

	for _, region := range r.regions {
		if region.Name() == pack.Name {
			region.Update(*pack)
			return nil
		}
	}

	return errors.New("No such region: " + pack.Name)
}

func (rs *RegionStatistics) RemoveRegion(region string) error {
	return rs.RemoveRegions([]string{region})
}

func (rs *RegionStatistics) RemoveRegions(regs []string) error {
	if regs == nil {
		return errors.New("Got nil RegionRemovePackage.")
	}

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

	return nil
}

// Generates a RegionStatsMap and
// sends it away on the publish channel.
func (rs RegionStatistics) generate(publisher chan<- RegionStatsMap) {
	publisher <- rs.Generate()
}

// inRange determines if the two coordinates are within range
func inRange(p1, p2 gr.XYer, distance float64) bool {
	l := math.Sqrt(math.Pow(p1.X()-p2.X(), 2) + math.Pow(p1.Y()-p2.Y(), 2))
	return l <= distance
}

// Function used for calculating new fixation points given a new data point
func newFixation(p1, p2 gr.XYer, numPoints int) *gr.Point2D {
	dx := p2.X() - p1.X()
	dy := p2.Y() - p1.Y()

	return &gr.Point2D{
		Xf: p1.X() + dx/float64(numPoints),
		Yf: p1.Y() + dy/float64(numPoints),
	}
}

func (rs RegionStatistics) Generate() RegionStatsMap {
	stats := make([]RegionStatInfo, len(rs.regions))
	prevTime := make([]*time.Time, len(stats)) // The last time stamp within the region

	prevFixInRegion := make([]bool, len(rs.regions))

	// Fixation init
	currFixation := &gr.Point2D{}
	currFixation = nil
	prevFixation := &gr.Point2D{}
	prevFixation = nil

	fixationRange := 0.05
	coordsInFixation := 0
	isNewFixation := true

	for coord := range rs.getCoords() { // Alot of coords
		if currFixation == nil {
			// First data coordinate
			currFixation = &coord.Filtered
			coordsInFixation = 1
			isNewFixation = false
		} else if inRange(currFixation, coord.Filtered, fixationRange) {
			// Update currFixation
			coordsInFixation++
			currFixation = newFixation(currFixation, coord.Filtered, coordsInFixation)
			isNewFixation = false
		} else { // Not in range, new fixation, set prevFixation
			coordsInFixation = 1
			prevFixation = currFixation
			currFixation = &coord.Filtered
			isNewFixation = true
		}

		for i, r := range rs.regions {
			// Overhead
			if isNewFixation {
				prevFixInRegion[i] = r.Contains(prevFixation)
			}

			if r.Contains(currFixation) && prevTime[i] == nil { // Enter
				prevTime[i] = &coord.Timestamp
				if !prevFixInRegion[i] { // Normal enter
					stats[i].Looks++
				}
				// If the previous fixation was in the region, it counts as a re-enter
				// and skipping the incrementation of looks

			} else if r.Contains(currFixation) && prevTime[i] != nil { // Inside
				stats[i].TimeInside += InsideTime(coord.Timestamp.Sub(*prevTime[i]))
				prevTime[i] = &coord.Timestamp

			} else if !r.Contains(currFixation) && prevTime[i] != nil { // Leave
				stats[i].TimeInside += InsideTime(coord.Timestamp.Sub(*prevTime[i]))
				if isNewFixation {
					prevTime[i] = nil
				} else { // If it was not a new "jump" this coordinate is on the border of the region and should count
					prevTime[i] = &coord.Timestamp
				}
			}
		}
	}

	var retMap = make(RegionStatsMap)

	for i, r := range rs.regions {
		retMap[r.Name()] = stats[i]
	}

	return retMap
}

// Start calls the Start method of coordhandler, enabling the collection of data
func (rs RegionStatistics) Start() {
	rs.coordinates.Start()
}

// Stop calls the Stop method of coordhandler, disabling the collection of data
func (rs RegionStatistics) Stop() {
	rs.coordinates.Stop()
}

func (rs RegionStatistics) Restart() {
	rs.Flush()
	//rs.Logger().Println("Flushing region stats buffer")
}

// Restart calls the Restart method of coordhandler, flushing the collection of data
func (rs RegionStatistics) Flush() {
	rs.coordinates.Flush()
}

type RegionStatsMap map[string]RegionStatInfo

//TODO rename
type RegionStatInfo struct {
	Looks      int        `json:"looks"`
	TimeInside InsideTime `json:"time"`
}

type RegionStats map[string]RegionStats

type InsideTime time.Duration

func (it InsideTime) MarshalText() ([]byte, error) {
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
