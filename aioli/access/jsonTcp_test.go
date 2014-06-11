package access

import (
	"testing"
	//"time"

	//"github.com/maxnordlund/breamio/aioli"
)

func TestTCPServer_Registration(t *testing.T) {
	registerTCPJSON()
	if _, ok := servers["TCP(JSON)"]; !ok {
		t.Error("Server is not registred.")
	}
}
