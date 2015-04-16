package access

import (
	"testing"
)

func TestTCPServer_Registration(t *testing.T) {
	registerTCPJSON()
	if _, ok := servers["TCP(JSON)"]; !ok {
		t.Error("Server is not registred.")
	}
}
