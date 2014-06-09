package configLoader

import (
	bl "github.com/maxnordlund/breamio/beenleigh"
	"io"
	"log"
	"os"
)

var logger = log.New(os.Stdout, "[ConfigLoader]", log.LstdFlags)

func init() {
	bl.Register(bl.NewRunHandler(Run))
}

type ConfigRun struct {
}

func (cr *ConfigRun) Run(logic bl.Logic, closer <-chan struct{}){
	logger.Println("Initilizing Config Loader.")
	newCh := logic.RootEmitter().Subscribe("new:configloader", bl.Spec{}).(<-chan bl.Spec)
	defer logic.RootEmitter().Unsubscribe("new:configloader", newCh)

	for {
		select {
			case event := <-newCh:
				New(logic, event)
			case <-closer: return
		}
	}
}
//New creates a new config loader that loads the local file specified in the provided specification
//
//
func New(logic bl.Logic, spec bl.Spec){
	filename := spec.Data
	// Read the file and send on the logic event emitter
}
