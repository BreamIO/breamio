package access

import (
	"fmt"
	"github.com/maxnordlund/breamio/module"

	"github.com/maxnordlund/breamio/aioli"
	"github.com/maxnordlund/breamio/beenleigh"
)

type AccessServer interface {
	Listen(aioli.IOManager, module.Logger)
}

var servers = make(map[string]AccessServer)
var ioman aioli.IOManager

func Register(name string, as AccessServer) {
	servers[name] = as
}

func init() {
	beenleigh.Register(&AioliAccess{})
}

func GetIOManager() aioli.IOManager {
	return ioman
}

type AioliAccess struct {
	ioman aioli.IOManager
}

func (AioliAccess) String() string {
	return "AioliAccess"
}

func (AioliAccess) New(module.Constructor) module.Module {
	return module.Dummy
}

func (aa *AioliAccess) Run(logic beenleigh.Logic) {
	ioman = aioli.New(logic, beenleigh.NewLoggerS("Aioli"))
	go ioman.Run()

	for name, as := range servers {
		logger := beenleigh.NewLoggerS(fmt.Sprintf("AioliAccess (%s)", name))
		go as.Listen(ioman, logger)
	}
}

func (aa *AioliAccess) Close() error {
	return aa.ioman.Close()
}
