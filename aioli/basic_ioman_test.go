package aioli

import (
	//"log"
	//"sync"
	//"bytes"
	//"time"
	"testing"
	"github.com/maxnordlund/breamio/briee"
)

func TestIoman(t *testing.T) {
	ee := briee.NewEventEmitter()
	// ee.Run()
	ioman := NewIOManager()
	
	// Add event emitter
	err := ioman.AddEE(&ee, 1)
	if err != nil {
		t.Errorf("Unable to add event emitter")
	}

	// Remove just added event emitter
	err = ioman.RemoveEE(1)
	if err != nil {
		t.Errorf("Unable to remove event emitter")
	}

	dataCh := make(chan ExtPkg)

	go ioman.Listen(dataCh)

	dataPkg := ExtPkg{"data event", 0, make([]byte, 10)}
	dataCh <- dataPkg
}
