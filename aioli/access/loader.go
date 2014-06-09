package access

import (
	"github.com/maxnordlund/breamio/aioli"
	"io/ioutil"
	"log"
	"os"
)

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

	file, err := os.Open("config.json")
	if err != nil {
		logger.Fatal("Unable to open file")
	}

	codec := aioli.NewCodec(file)
	go ioman.Listen(codec, logger)
}
