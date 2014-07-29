package director

import (
	"github.com/maxnordlund/breamio/beenleigh"
	"github.com/maxnordlund/breamio/briee"
	"reflect"
)

func init() {
	beenleigh.Register(beenleigh.NewRunHandler(runner))
}

func runner(logic beenleigh.Logic, closer <-chan struct{}) {
	briee.RegisterGlobalEventType("drawer:settings", reflect.TypeOf(DrawerSettings{}))
}

type DrawerSettings struct {
	Radius    uint
	Thickness uint
	Alpha     float64
	Color     []uint
}
