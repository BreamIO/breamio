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

type testStruct struct {
	Key string
}

func TestLoaderWithBL(t *testing.T) {
	if _, ok := servers["ConfigLoader"]; !ok {
		t.Error("Server is not registred.")
	}
	bl := been.New(briee.New)
	defer bl.Close()
	re := bl.RootEmitter()
	go bl.ListenAndServe()
	sub := re.Subscribe("testEvent1", testStruct{}).(<-chan testStruct)
	registerLoader()
	data := <-sub
	if data.Key != "First" {
		t.Fatal("Wrong")
	}
	data = <-sub
	if data.Key != "Second" {
		t.Fatal("Wrong")
	}

	//t.Logf("%v\n",data)
}
