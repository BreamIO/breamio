package director

import (
	"github.com/maxnordlund/breamio/briee"
	"github.com/maxnordlund/breamio/moduler"
	"reflect"
)

func init() {
	moduler.Register(Director{})
}

type Director struct{}

func (Director) String() string {
	return "Director"
}

func (Director) Run(logic moduler.Logic) {
	briee.RegisterGlobalEventType("drawer:settings", reflect.TypeOf(DrawerSettings{}))
}

func (Director) New(moduler.Constructor) moduler.Module {
	return moduler.Dummy
}

type DrawerSettings struct {
	Radius    uint
	Thickness uint
	Alpha     float64
	Color     []uint
	MaxLength uint
}
