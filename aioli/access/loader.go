package access

import (
	"bytes"
	"encoding/json"
	"github.com/maxnordlund/breamio/aioli"
	"io/ioutil"
	"log"
)

// LoaderPkg is a ExtPkg but with readable Data
type LoaderPkg struct {
	Event     string
	ID        int
	Data      interface{}
}

// This is the structure the config file has
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
	content, _ := ioutil.ReadFile("config.json") // TODO Constant filename

	// Unmarshal the file content into a MultiLoaderPkg
	var pkgSlice MultiLoaderPkg
	json.Unmarshal(content, &pkgSlice)

	events := make([]byte, 0)

	for _, pkgObj := range pkgSlice.Events {
		dataField := []byte(pkgObj.Data.(string)) // Parse Data field as a byte

		extPkg := aioli.ExtPkg{
			Event:     pkgObj.Event,
			ID:        pkgObj.ID,
			Data:      dataField,
		}

		byteExtPkg, err := json.Marshal(extPkg)
		if err != nil {
			logger.Print(err)
		}
		events = append(events, byteExtPkg...) // Appending the encoded ExtPkgs
	}

	buf := bytes.NewBuffer(events)
	codec := aioli.NewCodec(buf)
	go ioman.Listen(codec, logger)
}
