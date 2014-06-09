package access

import (
	"testing"
	"github.com/maxnordlund/breamio/briee"
	been "github.com/maxnordlund/breamio/beenleigh"
)


func TestLoaderRegistration(t *testing.T){
	registerLoader()
	if _, ok := servers["ConfigLoader"]; !ok {
		t.Error("Server is not registred.")
	}
}

func TestLoaderWithBL(t *testing.T){
	//registerLoader()
	if _, ok := servers["ConfigLoader"]; !ok {
		t.Error("Server is not registred.")
	}
	bl := been.New(briee.New)
	bl.ListenAndServe()
}
