package access

import (
	"log"
	"os"

	"github.com/maxnordlund/breamio/aioli"
	"github.com/maxnordlund/breamio/beenleigh"
)

type AccessServer interface {
	Listen(aioli.IOManager, *log.Logger)
}

var servers = make(map[string]AccessServer)
var ioman aioli.IOManager

func Register(name string, as AccessServer) {
	servers[name] = as
}

func init() {
	beenleigh.Register(beenleigh.NewRunHandler(accessRun))
}

func GetIOManager() aioli.IOManager {
	return ioman
}

func accessRun(logic beenleigh.Logic, closeCh <-chan struct{}) {
	ioman = aioli.New(logic)
	log.Println("Starting ExternalAccessService")
	go ioman.Run()
	go func() {
		<-closeCh
		ioman.Close()
	}()

	for name, as := range servers {
		logger := log.New(os.Stdout, "["+name+"] ", log.LstdFlags)
		go as.Listen(ioman, logger)
	}
}
