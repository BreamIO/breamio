package access

import (
	"fmt"

	"github.com/maxnordlund/breamio/moduler"
	"github.com/maxnordlund/breamio/remote"
)

type AccessServer interface {
	Listen(remote.IOManager, moduler.Logger)
}

var servers = make(map[string]AccessServer)
var remoteman remote.IOManager

func Register(name string, as AccessServer) {
	servers[name] = as
}

func init() {
	moduler.Register(&AremoteliAccess{})
}

func GetIOManager() remote.IOManager {
	return remoteman
}

type AremoteliAccess struct {
	remoteman remote.IOManager
}

func (AremoteliAccess) String() string {
	return "AremoteliAccess"
}

func (AremoteliAccess) New(moduler.Constructor) moduler.Module {
	return moduler.Dummy
}

func (aa *AremoteliAccess) Run(logic moduler.Logic) {
	remoteman = remote.New(logic, moduler.NewLoggerS("Aremoteli"))
	go remoteman.Run()

	for name, as := range servers {
		logger := moduler.NewLoggerS(fmt.Sprintf("AremoteliAccess (%s)", name))
		go as.Listen(remoteman, logger)
	}
}

func (aa *AremoteliAccess) Close() error {
	return aa.remoteman.Close()
}
