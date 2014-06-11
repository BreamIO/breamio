package access

import (
	"encoding/json"
	"github.com/maxnordlund/breamio/aioli"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"
)

// LoaderPkg is a ExtPkg but with readable Data
type LoaderPkg struct {
	Event string
	ID    int
	Data  interface{}
}

// This is the structure the config file has
type MultiLoaderPkg struct {
	Events []LoaderPkg
}

var Configfile = "config.json"

func init() {
	if filepath := os.Getenv("EYESTREAM"); filepath != "" {
		Configfile = path.Join(filepath, Configfile)
	}
	registerLoader()
}

func registerLoader() {
	Register("ConfigLoader", ConfigLoader{})
}

type ConfigLoader struct{}

func (cl ConfigLoader) Listen(ioman aioli.IOManager, logger *log.Logger) {
	content, _ := ioutil.ReadFile(Configfile)

	// Unmarshal the file content into a MultiLoaderPkg
	var pkgSlice MultiLoaderPkg
	json.Unmarshal(content, &pkgSlice)

	for _, pkgObj := range pkgSlice.Events {
		dataField, err := json.Marshal(pkgObj.Data)
		if err != nil {
			logger.Fatal(err)
		}

		extPkg := aioli.ExtPkg{
			Event:     pkgObj.Event,
			Subscribe: false,
			ID:        pkgObj.ID,
			Data:      dataField,
			Error:     nil,
		}
		ioman.Dispatch(extPkg)
		time.Sleep(1000 * time.Millisecond)
	}
}
