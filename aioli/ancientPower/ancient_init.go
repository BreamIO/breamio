// +build ancient

package ancientPower

import (
	bl "github.com/maxnordlund/breamio/beenleigh"
)

func init() {
	bl.Register(&AncientRun{make(chan struct{})})
}

type AncientRun struct {
	closing chan struct{}
}

func (ar *AncientRun) Run(logic bl.Logic) {
	logger.Println("Initializing AncientPower")
	ar.closing = make(chan struct{})
	newCh := logic.RootEmitter().Subscribe("new:ancientpower", bl.Spec{}).(<-chan bl.Spec)
	defer logic.RootEmitter().Unsubscribe("new:ancientpower", newCh)

	for {
		select {
		case event := <-newCh:
			New(logic, event)
		case <-ar.closing:
			return
		}
	}

}

func (ar *AncientRun) Close() error {
	if ar.closing != nil {
		close(ar.closing)
	}
	return nil
}
