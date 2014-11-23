package beenleigh

import (
	"github.com/maxnordlund/breamio/briee"
	"reflect"
	"strings"
)

func RunFactory(l Logic, f Factory) {
	news := l.RootEmitter().Subscribe("new:"+f.String(), Spec{}).(<-chan Spec)
	defer l.RootEmitter().Unsubscribe("new:"+f.String(), news)
	for n := range news {
		// Would have prefered m as the logger object,
		// but until such time where I can call a method on a
		// object before creating it, I have to use the factory
		m := f.New(Constructor{
			Logger:     NewLogger(f),
			Parameters: n.Data,
		})
		m.Logger().Println("Booting up!")
		if runner, ok := m.(Runner); ok {
			go runner.Run(l)
		} else {
			go RunModule(l, n.Emitter, m)
		}
	}
}

var methodType = reflect.TypeOf(MethodEvent{})

type exportedMethod struct {
	name, event  string
	returnevents []string
}

func RunModule(l Logic, emitterId int, m Module) {
	typ := reflect.TypeOf(m)
	val := reflect.ValueOf(m)

	var exported []exportedMethod

	//Look for EventMethods among fields
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.Type == methodType {
			//Evented method declaration
			//Figure out method name from field name and tag
			//Store in structure to be iterated later

			name := field.Tag.Get("method")
			if name == "" {
				//Use heuristic and field name
				name = strings.TrimPrefix(field.Name, "Method")
				if name == "" {
					name = "Method" //If all else fails.
				}
			}

			event := field.Tag.Get("event")
			if event == "" {
				//Use heuristic
				event = m.String() + ":" + name
			}

			rets := strings.Split(field.Tag.Get("returns"), ",")
			returns := make([]string, 0, len(rets))
			for _, ret := range rets {
				if ret == "_" {
					returns = append(returns, "_")
				} else {
					returns = append(returns, m.String()+":"+ret)
				}
			}
			exported = append(exported, exportedMethod{name, event, returns})
		}
	}

	emitter := l.CreateEmitter(emitterId)

	for _, em := range exported {
		methodType, ok := typ.MethodByName(em.name)
		if !ok {
			l.Logger().Panicf("Method %s on %s does not exist.", em.name, typ.Name())
		}

		method := val.MethodByName(em.name)

		if suitable(methodType) {
			go RunMethod(method, em, emitter, l.Logger())
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

func RunMethod(method reflect.Value, em exportedMethod, emitter briee.EventEmitter, l Logger) {
	t := reflect.New(method.Type().In(0))
	ch := emitter.Subscribe(em.event, t)
	defer emitter.Unsubscribe(em.event, ch)
	if len(em.returnevents) > method.Type().NumOut() {
		l.Panicf("More return events than return values.")
	}

	returns := make([]reflect.Value, method.Type().NumOut())
	for i, retevent := range em.returnevents {
		if retevent == "_" {
			continue
		}
		t := reflect.New(method.Type().Out(i))
		rch := emitter.Publish(retevent, t)
		returns[i] = reflect.ValueOf(rch)
	}

	vch := reflect.ValueOf(ch)
	for {
		val, ok := vch.Recv()
		if !ok {
			return //Event is closed.
		}
		rets := method.Call([]reflect.Value{val})
		//Send return values to their respective
		for i, val := range rets {
			if returns[i].IsValid() {
				returns[i].Send(val)
			}
		}
	}
}
