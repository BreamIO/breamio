package access

import (
	"encoding/json"
	"github.com/maxnordlund/breamio/aioli"
	"io/ioutil"
	"log"
	//"os"
	"bytes"
	"strings"
)

type LoaderPkg struct {
	Event     string // Name of the event
	Subscribe bool   // Should the handler setup a subscription channel for this event and client.
	ID        int    // Event Emitter identifier, 0 for broadcast
	//Data      interface{} `json:",string"` // Encoded data of the underlying struct for the event.
	Data interface{}
	Error *aioli.Error // Meta-data to indicate errors in requests.
}

type MultiLoaderPkg struct {
	Events []LoaderPkg
}

func init() {
	registerLoader()
}

func registerLoader() {
	Register("ConfigLoader", ConfigLoader{})
}

type ConfigLoader struct{}

func (cl ConfigLoader) Listen(ioman aioli.IOManager, logger *log.Logger) {
	content, _ := ioutil.ReadFile("config.json")
	logger.Println(string(content))

	// Unmarshal file content
	var pkgSlice MultiLoaderPkg
	json.Unmarshal(content, &pkgSlice)
	logger.Println(pkgSlice)

	events := make([]string, len(pkgSlice.Events))

	// Check the content of the slice, OK!
	for i, pkgObj := range pkgSlice.Events {
		logger.Println(pkgObj)

		// Marshal the data field
		dataField, _ := json.Marshal(pkgObj.Data)

		extPkg := aioli.ExtPkg {
			Event: pkgObj.Event,
			Subscribe: pkgObj.Subscribe,
			ID: pkgObj.ID,
			Data: dataField,
			Error: pkgObj.Error,
		}

		byteExtPkg,_ := json.Marshal(extPkg)
		events[i] = string(byteExtPkg)
	}
	buf := bytes.NewBuffer([]byte(strings.Join(events, "")))
	codec := aioli.NewCodec(buf)
	go ioman.Listen(codec, logger)
}
