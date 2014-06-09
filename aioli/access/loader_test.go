package access

import (
	been "github.com/maxnordlund/breamio/beenleigh"
	"github.com/maxnordlund/breamio/briee"
	"testing"
)

func TestLoaderRegistration(t *testing.T) {
	registerLoader()
	if _, ok := servers["ConfigLoader"]; !ok {
		t.Error("Server is not registred.")
	}
}

type testStruct struct{
	Key int
}

func TestLoaderWithBL(t *testing.T) {
	// This test is broken
	if _, ok := servers["ConfigLoader"]; !ok {
		t.Error("Server is not registred.")
	}
	bl := been.New(briee.New)
	re := bl.RootEmitter()
	go bl.ListenAndServe()
	sub := re.Subscribe("testEvent1", testStruct{}).(<-chan testStruct)
	//registerLoader()
	data := <-sub
	t.Logf("%v\n",data)
	bl.Close()
}
