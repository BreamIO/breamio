package beenleigh_test

import (
	bl "github.com/maxnordlund/breamio/beenleigh"
	"github.com/maxnordlund/breamio/briee"
	"testing"
	"time"
)

func TestRunHandle_Run(t *testing.T) {
	test := false
	rc := bl.NewRunHandler(func(logic bl.Logic, closeCh <-chan struct{}) {
		test = true;
	})
	rc.Run(bl.New(briee.New))
	if !test {
		t.Error("Function did not run.")
	}
}

func TestRunHandle_Close(t *testing.T) {
	test := false
	rc := bl.NewRunHandler(func(logic bl.Logic, closeCh <-chan struct{}) {
		<-closeCh
		test = true;
	})
	go rc.Run(bl.New(briee.New))
	rc.Close()
	time.Sleep(time.Millisecond)
	if !test {
		t.Error("Function did not run.")
	}
}