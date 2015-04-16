package access

import (
	"log"
	"testing"
	"time"

	"github.com/maxnordlund/breamio/briee"
	"github.com/maxnordlund/breamio/io"
	"github.com/maxnordlund/breamio/moduler"
)

func TestRegister(t *testing.T) {
	//Test server should no preexist.
	if _, ok := servers["test"]; ok {
		t.Error("Why does a test server alrady exist?")
	}

	Register("test", &TestServer{})

	//Register should store the server in the map
	if _, ok := servers["test"]; !ok {
		t.Error("Test server was not stored in map!")
	}
}

func TestAccessRun(t *testing.T) {
	servers = make(map[string]AccessServer) // Reset servers.
	bl := moduler.New(briee.New)
	ts := &TestServer{}
	Register("test", ts)
	cCh := make(chan struct{})
	accessRun(bl, cCh)
	time.Sleep(time.Millisecond)
	if !ts.started {
		t.Error("Server not started.")
	}
	if ts.manager == nil {
		t.Error("Nil IOManager given to server.")
	}
}

type TestServer struct {
	started bool
	manager io.IOManager
}

func (ts *TestServer) Listen(ioman io.IOManager, l *log.Logger) {
	ts.started = true
	ts.manager = ioman
}
