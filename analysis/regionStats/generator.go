package regionStats

import (
	"errors"
	"log"
	"os"
	"strconv"
	"time"
	"math"

	"github.com/maxnordlund/breamio/analysis"
	"github.com/maxnordlund/breamio/beenleigh"
	"github.com/maxnordlund/breamio/briee"
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
	beenleigh.Register(&RegionRun{
		generators: make(map[int]*RegionStatistics),
		closeChan:  make(chan struct{}),
	})
}

// RegionRun creates and runs generators,
// and terminates them once closed.
type RegionRun struct {
	generators map[int]*RegionStatistics
	closeChan  chan struct{}
}

func (r *RegionRun) Run(logic beenleigh.Logic) {
	log := log.New(os.Stderr, "[ RegionRun ] ", log.LstdFlags)
	log.Println("Registering in EE")
	var newChan <-chan *Config
	var ee briee.EventEmitter

	ee = logic.RootEmitter()
	newChan = ee.Subscribe("new:regionStats", new(Config)).(<-chan *Config)
	defer ee.Unsubscribe("new:regionStats", newChan)
	for {
		select {
		case rc := <-newChan:
			log.Println("Starting a new generator for emitter:", rc.Emitter)
			r.generators[rc.Emitter] =
				New(logic.CreateEmitter(rc.Emitter), rc.Duration, rc.Hertz, rc.GenerationInterval)
		case <-r.closeChan:
			log.Println("Shutting down")
			break
		}
	}
}

func (r *RegionRun) Close() error {
	close(r.closeChan)

	for _, generator := range r.generators {
		close(generator.closeChan)
	}

	return nil
}

type RegionStatistics struct {
	coordinates        *analysis.CoordBuffer
	regions            []Region
	publish            chan<- RegionStatsMap
	closeChan          chan struct{}
	generationInterval time.Duration
}

func New(ee briee.PublishSubscriber, dur string, hertz uint, genIntv string) *RegionStatistics {
	log := log.New(os.Stderr, "[ RegionStats ] ", log.LstdFlags)

	duration, err := time.ParseDuration(dur)
	if err != nil {
		log.Println(err, "Defaulting duration to 60 seconds")
		duration = time.Minute
	}
	generationInterval, err := time.ParseDuration(genIntv)
	if err != nil {
		log.Println(err, "Defaulting generation interval to 60 seconds")
		duration = time.Minute
	}

	ch := ee.Subscribe("tracker:etdata", &gr.ETData{}).(<-chan *gr.ETData)

	addRegionCh := ee.Subscribe("regionStats:addRegion", new(RegionDefinitionPackage)).(<-chan *RegionDefinitionPackage)
	updateRegionCh := ee.Subscribe("regionStats:updateRegion", new(RegionUpdatePackage)).(<-chan *RegionUpdatePackage)
	removeRegionCh := ee.Subscribe("regionStats:removeRegion", make([]string, 0, 0)).(<-chan []string)

	startch := ee.Subscribe("regionStats:start", struct{}{}).(<-chan struct{})
	stopch := ee.Subscribe("regionStats:stop", struct{}{}).(<-chan struct{})
	restartch := ee.Subscribe("regionStats:restart", struct{}{}).(<-chan struct{})

	rs := &RegionStatistics{
		coordinates:        analysis.NewCoordBuffer(ch, duration, hertz),
		regions:            make([]Region, 0),
		publish:            ee.Publish("regionStats:regions", make(RegionStatsMap)).(chan<- RegionStatsMap),
		closeChan:          make(chan struct{}),
		generationInterval: generationInterval,
	}

	go func(rs *RegionStatistics) {
		defer func() {
			close(rs.publish)
			ee.Unsubscribe("regionStats:addRegion", addRegionCh)
			ee.Unsubscribe("regionStats:updateRegion", updateRegionCh)
			ee.Unsubscribe("regionStats:removeRegion", removeRegionCh)
			ee.Unsubscribe("regionStats:start", startch)
			ee.Unsubscribe("regionStats:stop", stopch)
			ee.Unsubscribe("regionStats:restart", restartch)
			ee.Unsubscribe("tracker:etdata", ch)
		}()
		for {
			select {
			case <-rs.closeChan:
				log.Println("Closing down regionStatistics")
				return

			// TODO refactor select case structure
			case regionDefPak, ok := <-addRegionCh:
				if !ok {
					return
				}
				log.Println("Adding region: ", regionDefPak.Name)
				err := rs.addRegion(regionDefPak)

				if err != nil {
					log.Println(err.Error())
				}

			case regionUpdate, ok := <-updateRegionCh:
				if !ok {
					return
				}
				log.Println("Updating region: ", regionUpdate.Name)
				err := rs.updateRegion(regionUpdate)

				if err != nil {
					log.Println(err.Error())
				}

			case regs, ok := <-removeRegionCh:
				if !ok {
					return
				}
				log.Println("Removing region(s):", regs)
				err := rs.removeRegions(regs)

				if err != nil {
					log.Println(err.Error())
				}

			// start
			case _, ok := <-startch:
				if !ok {
					return
				}
				rs.Start()
				log.Println("Starting region stats buffer")

			// stop
			case _, ok := <-stopch:
				if !ok {
					return
				}
				rs.Stop()
				log.Println("Stopping region stats buffer")

			// flush
			case _, ok := <-restartch:
				if !ok {
					return
				}
				rs.Flush()
				log.Println("Flushing region stats buffer")

			case <-time.After(rs.generationInterval):
				rs.Generate()
			}
		}
	}(rs)

	return rs
}

func (rs RegionStatistics) getCoords() (coords chan *gr.ETData) {
	return rs.coordinates.GetCoords()
}

func (rs *RegionStatistics) addRegion(pack *RegionDefinitionPackage) error {
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

func (r *RegionStatistics) updateRegion(pack *RegionUpdatePackage) error {
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

func (rs *RegionStatistics) removeRegions(regs []string) error {
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
func (rs RegionStatistics) Generate() {
	rs.publish <- rs.generate()
}

// inRange determines if the two coordinates are within range
func inRange(p1, p2 gr.XYer, distance float64) bool {
	l := math.Sqrt(math.Pow(p1.X() - p2.X(), 2) + math.Pow(p1.Y() -p2.Y(), 2))
	return l <= distance
}

// Function used for calculating new fixation points given a new data point
func newFixation(p1, p2 gr.XYer, numPoints int) *gr.Point2D {
	dx := p2.X() - p1.X()
	dy := p2.Y() - p1.Y()

	return &gr.Point2D {
		Xf: p1.X() + dx/float64(numPoints),
		Yf: p1.Y() + dy/float64(numPoints),
	}
}

func (rs RegionStatistics) generate() RegionStatsMap {
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

	for coord := range rs.getCoords() {        // Alot of coords
		if currFixation == nil {
			// First data coordinate
			currFixation = &coord.Filtered
			coordsInFixation = 1
			isNewFixation = false
		} else if inRange(currFixation, coord.Filtered, fixationRange){
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
				prevFixInRegion[i] = rs.Contains(prevFixation)
			}

			if rs.Contains(currFixation) && prevTime[i] == nil { // Enter
				prevTime[i] = &coord.Timestamp
				if !prevFixationInRegion[i] { // Normal enter
					stats[i].Looks++
				}
				// If the previous fixation was in the region, it counts as a re-enter
				// and skipping the incrementation of looks

			} else if rs.Contains(currFixation) && prevTime[i] != nil { // Inside
				stats[i].TimeInside += InsideTime(coord.Timestamp.Sub(*prevTime[i]))
				prevTime[i] = &coord.Timestamp

			} else if rs.Contains(currFixation) && prevTime[i] == nil { // Leave
				stats[i].TimeInside += InsideTime(coord.Timestamp.Sub(*prevTime[i]))
				prevTime[i] = nil
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
