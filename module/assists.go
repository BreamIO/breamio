package module

var Dummy = Module(dummymodule{SimpleModule{"DUMMY", nil}})

type dummymodule struct {
	SimpleModule
}

/*
Assistive type to aid in creation of modules that are intended to operate as a singleton.
Name is mandatory, but if Module is defined, there is no need for Constructor and vice-versa.
The factory will manufacture a module if no known instance is already available.
*/
type SingletonFactory struct {
	Name        string
	Constructor func(Constructor) Module
	Module
}

func (sf SingletonFactory) String() string {
	return sf.Name
}

func (sf SingletonFactory) New(c Constructor) Module {
	if sf.Module == nil {
		sf.Module = sf.Constructor(c)
	}
	return sf.Module
}
