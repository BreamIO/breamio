package access

import (
	"encoding/json"
	"encoding/base64"
	"github.com/maxnordlund/breamio/aioli"
	"io/ioutil"
	"log"
	//"os"
	"bytes"
	//"strings"
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
	//logger.Println(string(content))

	// Unmarshal file content
	var pkgSlice MultiLoaderPkg
	json.Unmarshal(content, &pkgSlice)
	//logger.Println(pkgSlice)

	//events := make([]string, len(pkgSlice.Events))
	//events := make([][]byte, len(pkgSlice.Events))
	events := make([]byte, 0)

	// Check the content of the slice, OK!
	for _, pkgObj := range pkgSlice.Events {
		//logger.Println(pkgObj)

		// Marshal the data field
		dataFieldRaw, err := json.Marshal(pkgObj.Data)
		if err != nil {
			logger.Print(err)
		}
		dataField := []byte(base64.StdEncoding.EncodeToString(dataFieldRaw))

		extPkg := aioli.ExtPkg {
			Event: pkgObj.Event,
			Subscribe: pkgObj.Subscribe,
			ID: pkgObj.ID,
			Data: dataField,
			Error: pkgObj.Error,
		}

		byteExtPkg, err := json.Marshal(extPkg)
		if err != nil {
			logger.Print(err)
		}
		// [TESTING START]
		logger.Print("Testing start")

		var tmpExtPkg aioli.ExtPkg
		err = json.Unmarshal(byteExtPkg, &tmpExtPkg)
		logger.Print(byteExtPkg)
		logger.Print(tmpExtPkg)

		if err != nil {
			logger.Print(err)
		}

		tmpstr, err := base64.StdEncoding.DecodeString(string(tmpExtPkg.Data))
		// TODO Try to find out what is going wrong!!
		//tmpdata := make([]byte, 256)
		err = json.Unmarshal(tmpExtPkg.Data, &tmpdata)
		if err != nil {
			logger.Print(err)
		}

		logger.Print(tmpdata)

		logger.Print("Testing end")
		// [TESTING END]
		//events[i] = byteExtPkg
		events = append(events, byteExtPkg...)
	}

	//buf := bytes.NewBuffer([]byte(strings.Join(events, "")))
	buf := bytes.NewBuffer(events)
	logger.Println(buf.String())
	codec := aioli.NewCodec(buf)
	go ioman.Listen(codec, logger)
}
