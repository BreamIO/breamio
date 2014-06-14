// Config File Maker (CFM)
//
// This program creates a config file for further use with configLoader.
// First command line argument is the name of the input CFM config file
// Second command line argument is the name of the output config file
//
// The CFM config file declares the eye trackers needed,
// a template statistic definition and the regions that should be
// created for each tracker.
//
// See cfm_example_config.json for an example use
// It will generate the file example_config_output.json
//
// Example use:
//
// $> cfm cfm_example_config.json example_config_ouput.json

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"flag"
)

type LoaderPkg struct {
	Event string
	ID    int
	Data  interface{}
}

// This is the structure the config file has
type MultiLoaderPkg struct {
	Events []LoaderPkg
}

func main() {
	// Parse input flags
	inputfilePtr := flag.String("input", "cfm_input.json", "Config File Maker input file, example use see \"cfm_example_input.json\"")
	outputfilePtr := flag.String("output", "config_output.json", "Config File Maker output file, example use see \"cfm_example_output.json\"")
	flag.Parse()
	inputfile := *inputfilePtr
	outputfile := *outputfilePtr

	if inputfile == outputfile {
		log.Println("Cant use the same input file as output file.")
		os.Exit(1)
	}
	// Read content
	content, err := ioutil.ReadFile(inputfile)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	// Unmarshal into a map[string]interface{}
	var segments map[string]interface{}
	err = json.Unmarshal(content, &segments)
	if err != nil {
		log.Println(err)
	}

	trackers := segments["Trackers"].([]interface{})
	regionTmpl := segments["RegionStatsTemplate"].([]interface{})[0]
	regionDefs := segments["RegionDefinitions"].([]interface{})

	events := make([]LoaderPkg, 0)
	ids := make([]int, 0)

	// Add the trackers configurations
	for _, tracker := range trackers {
		tmap := tracker.(map[string]interface{})

		// Extract the ID from the Data field
		id := int(tmap["Data"].(map[string]interface{})["Emitter"].(float64))
		ids = append(ids, id)

		event := LoaderPkg{
			Event: tmap["Event"].(string),
			ID:    int(tmap["ID"].(float64)),
			Data:  tmap["Data"],
		}

		events = append(events, event)
	}

	// Add stats for each tracker
	for _, id := range ids {
		tmplmap := regionTmpl.(map[string]interface{})
		// Need a copy of the data map in order to change the Emitter value
		datacopy := make(map[string]interface{})

		for k, v := range tmplmap["Data"].(map[string]interface{}) {
			datacopy[k] = v
		}

		datacopy["Emitter"] = id

		event := LoaderPkg{
			Event: tmplmap["Event"].(string),
			ID:    int(tmplmap["ID"].(float64)),
			Data:  datacopy,
		}
		events = append(events, event)
	}

	// Add regions for each tracker/stats
	for _, id := range ids {
		for _, region := range regionDefs {
			regionMap := region.(map[string]interface{})
			regionMapCopy := make(map[string]interface{})
			for k, v := range regionMap {
				regionMapCopy[k] = v
			}
			regionMapCopy["ID"] = id

			event := LoaderPkg{
				Event: regionMapCopy["Event"].(string),
				ID:    regionMapCopy["ID"].(int),
				Data:  regionMapCopy["Data"],
			}
			events = append(events, event)

		}
	}

	// Write the events to file
	config := MultiLoaderPkg{Events: events}
	jsonConfig, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Println(err)
	}

	err = ioutil.WriteFile(outputfile, jsonConfig, 0644)
	if err != nil {
		log.Println(err)
	}
}
