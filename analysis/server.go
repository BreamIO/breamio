package statistics

// import (
// 	"encoding/json"
// 	"flag"
// 	"fmt"
// 	"os"
// 	//	"time"
// 	//"strings"
// )

// var (
// 	jsonPath              = flag.String("jp", "", "The path to the jason file.")
// 	screenRatio           = flag.String("sr", "", "The screen ratio that is going to be used.")
// 	heatDurationInterval  = flag.Uint("hd", 300, "Number of seconds in heatmap interval")
// 	statsDurationInterval = flag.Uint("sd", 60, "Number of seconds in stats interval")
// 	statsHertz            = flag.Int("hz", 60, "The frequency of ET-Data")
// 	ip                    = flag.String("ip", "localhost:3031", "The ip and the port to the ET. ")
// )

// type RegionInfo map[string]Region

// func main() {
// 	flag.Parse()
// 	//coordHandler := connectTo(*ip)

// 	//statisticsGenerator := makeStatisticsGenerator()
// 	//coordHandler.AddListener(statisticsGenerator.timeList) TODO

// 	//heatMapGenerator := makeHeatmapGenerator()
// 	//coordHandler.AddListener(heatMapGenerator.GetTimeList())

// 	//runWebServer(statisticsGenerator, heatmapGenerator)
// 	//time.Sleep(time.Second*10)
// 	//heatMapGenerator.Generate(200, 200)
// 	//fmt.Fprint(os.Stderr, statisticsGenerator.Generate())
// }

// /*
// func makeHeatmapGenerator() HeatMapHandler {
// 	return newHeatMapHandler(
// 		time.Duration(*heatDurationInterval)*time.Second,
// 		*statsHertz)
// }
// */

// func makeStatisticsGenerator() /* *RegionHandler */ {
// 	file, err := os.Open(*jsonPath)

// 	if err != nil {
// 		fmt.Fprintln(os.Stderr, "error opening json,", err)
// 		os.Exit(1)
// 	}

// 	dec := json.NewDecoder(file)
// 	var aspectMap = make(AspectMap)
// 	dec.Decode(&aspectMap)

// 	/*regionHandler := NewRegionHandler(time.Duration(*statsDurationInterval)*time.Second, *statsHertz)
// 	regionHandler.AddRegions(aspectMap[*screenRatio])
// 	*/
// 	/*enc := json.NewEncoder(os.Stdout)
// 	enc.Encode(regionHandler.Generate())*/

// 	//return regionHandler
// }
