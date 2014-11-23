package director

import (
	"github.com/maxnordlund/breamio/beenleigh"
	"github.com/maxnordlund/breamio/briee"
	"github.com/maxnordlund/breamio/module"
	"reflect"
)

func init() {
	beenleigh.Register(Director{})
}

type Director struct{}

func (Director) String() string {
	return "Director"
}

func (Director) Run(logic beenleigh.Logic) {
	briee.RegisterGlobalEventType("drawer:settings", reflect.TypeOf(DrawerSettings{}))
}

func (Director) New(module.Constructor) module.Module {
	return module.Dummy
}

type DrawerSettings struct {
	Radius    uint
	Thickness uint
	Alpha     float64
	Color     []uint
	MaxLength uint
}
