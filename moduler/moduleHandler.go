package moduler

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/maxnordlund/breamio/briee"
)

func RunFactory(l Logic, f Factory) {
	news := l.RootEmitter().Subscribe("new:"+f.String(), Spec{}).(<-chan Spec)
	defer l.RootEmitter().Unsubscribe("new:"+f.String(), news)
	if closer, ok := f.(io.Closer); ok {
		defer closer.Close()
	}
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
			go func(runner Runner) {
				runner.Run(l)
				if closer, ok := runner.(io.Closer); ok {
					closer.Close()
				}
			}(runner)
		} else {
			go RunModule(l, n.Emitter, m)
		}
	}
}

var methodType = reflect.TypeOf(EventMethod{})

type exportedMethod struct {
	name, event  string
	returnevents []string
}

func RunModule(l Logic, emitterId int, m Module) {
	typ := reflect.TypeOf(m)
	val := reflect.ValueOf(m)

	var exported []exportedMethod

	styp := typ

	if styp.Kind() == reflect.Ptr {
		styp = typ.Elem()
	}

	if styp.Kind() != reflect.Struct {
		panic(fmt.Sprintf("RunModule does not support %s of %s", typ.Kind(), typ.String()))
	}

	//Look for EventMethods among fields
	for i := 0; i < styp.NumField(); i++ {
		field := styp.Field(i)
		if field.Type == methodType {
			//Evented method declaration
			//Figure out method name from field name and tag
			//Store in structure to be iterated later

			name := field.Tag.Get("method")
			if name == "" {
				//Use heuristic and field name
				name = strings.TrimPrefix(field.Name, "Method")
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
		m.Logger().Println("Automatic export of", em.name, "on", em.event)
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

// Helper function to determine if a method signature is suitable for
// event exporting.
// Returns true iff the method takes a single or no arguments.
// In the future this might also check the argument and return types for serlializability aswell.
func suitable(m reflect.Method) bool {
	return m.Type.NumIn() <= 2
}

func returnable(m reflect.Method) bool {
	
	//TODO remove or so 
	
	//TODO implement me
	return false
}

type Signal *struct{}

var Pulse Signal = Signal(&struct{}{})

func RunMethod(method reflect.Value, em exportedMethod, emitter briee.EventEmitter, l Logger) {
	t := reflect.ValueOf(Pulse)
	if method.Type().NumIn() == 1 {
		//Argument availiable
		t = reflect.New(method.Type().In(0)).Elem()
	}
	// TODO commented code
	// l.Println(em.event, t)
	ch := emitter.Subscribe(em.event, t.Interface())
	defer emitter.Unsubscribe(em.event, ch)
	if len(em.returnevents) > method.Type().NumOut() {
		l.Panicf("More return events than return values.")
	}

	returns := make([]reflect.Value, method.Type().NumOut())
	for i, retevent := range em.returnevents {
		if retevent == "_" {
			l.Printf("Method (%s) does not want to export returnvalue %d", em.name, i)
			continue
		}
		t := reflect.New(method.Type().Out(i)).Elem()
		// TODO Commented code
		// l.Println(retevent, t)
		rch := emitter.Publish(retevent, t.Interface())
		returns[i] = reflect.ValueOf(rch)
	}

	vch := reflect.ValueOf(ch)
	for {
		val, ok := vch.Recv()
		if !ok {
			return //Event is closed.
		}
		var rets []reflect.Value
		if method.Type().NumIn() == 0 {
			// Only reciever as argument
			rets = method.Call([]reflect.Value{})
		} else if method.Type().NumIn() == 1 {
			// reciever + 1 argument
			rets = method.Call([]reflect.Value{val})
		} else {
			l.Panicf("Wrong amount of arguments %d to method %s. Expected 0 or 1.", method.Type().NumIn(), method.Type().Name())
		}
		//Send return values to their respective
		for i, val := range rets {
			//The nil check allows a method to give out different events depending on circumstances
			//A nil event is simply discarded. If a event with no meaning other that "it has happened" is wanted, use Signal or *Signal instead.
			if returns[i].IsValid() && !toBeSent(val) {
				returns[i].Send(val)
			}
		}
	}
}

func toBeSent(val reflect.Value) bool {
	switch val.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func, reflect.Interface:
		return val.IsNil()
	default:
		return true
	}
}
