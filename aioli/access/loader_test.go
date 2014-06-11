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
	String string
	Boolean bool
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
	if data.String != "First" {
		t.Fatal("Expected %v, got %v\n", "First", data.String)
	}
	if data.Boolean != true {
		t.Fatal("Expected %v, got %v\n", true, data.Boolean)
	}
	data = <-sub
	if data.String != "Second" {
		t.Fatal("Expected %v, got %v\n", "Second", data.String)
	}
	if data.Boolean != false {
		t.Fatal("Expected %v, got %v\n", false, data.Boolean)
	}

	//t.Logf("%v\n",data)
}
