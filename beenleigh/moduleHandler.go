package beenleigh

import (
	"github.com/maxnordlund/breamio/module"
	"reflect"
)

func Run(l Logic, m module.Module) {
	typ := reflect.TypeOf(m)

	//Look for EventMethods among fields

	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)
		if suitable(method) {
			//Use l to get emitter, and subscribe to event
			if returnable(method) {
				//Publish that on event
			}
		}
	}
}
