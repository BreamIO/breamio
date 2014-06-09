package access

import (
	"log"
	"os"
	"github.com/maxnordlund/breamio/aioli"
)

func init() {
	registerLoader()
}

func registerLoader() {
	Register("ConfigLoader", ConfigLoader{})
}

type ConfigLoader struct{}

func (cl ConfigLoader) Listen(ioman aioli.IOManager, logger *log.Logger) {
	file, err := os.Open("config.json")
	if err != nil {
		logger.Fatal("Unable to open file")
	}
	codec := aioli.NewCodec(file)
	go ioman.Listen(codec, logger)
}
