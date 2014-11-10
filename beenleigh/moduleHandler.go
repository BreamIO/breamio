package beenleigh

import (
	"github.com/maxnordlund/breamio/module"
	"reflect"
)

func RunFactory(l Logic, f module.Factory) {
	news := l.RootEmitter().Subscribe("new:"+f.String(), map[string]interface{}{}).(<-chan map[string]interface{})
	defer l.RootEmitter().Unsubscribe("new:"+f.String(), news)
	for n := range news {
		// Would have prefered m as the logger object,
		// but until such time where I can call a method on a
		// object before creating it, I have to use the factory
		m := f.New(module.Constructor{
			Logger:     NewLogger(f),
			Parameters: n,
		})
		if runner, ok := m.(RunCloser); ok {
			go runner.Run(l)
		} else {
			go RunModule(l, m)
		}
	}
}

var methodType = reflect.TypeOf(MethodEvent{})

type exportedMethod struct {
	name, event  string
	returnevents []string
}

func RunModule(l Logic, m module.Module) {
	typ := reflect.TypeOf(m)

	var exported []exportedMethod

	//Look for EventMethods among fields
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.Type == methodType {
			//Evented method declaration
			//Figure out method name from field name and tag
			//Store in structure to be iterated later
		}
	}

	for _, em := range exported {
		method, ok := typ.MethodByName(em.name)
		if !ok {
			l.Logger().Panicf("Method %s on %s does not exist.", em.name, typ.Name())
		}

		if suitable(method) {
			//Use l to get emitter, and subscribe to event
			if returnable(method) {
				//Publish that on event
			}
		}
	}
}

func suitable(m reflect.Method) bool {
	//TODO implement me
	return false
}

func returnable(m reflect.Method) bool {
	//TODO implement me
	return false
}
