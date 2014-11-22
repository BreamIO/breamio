package module

var Dummy = Module(dummymodule{SimpleModule{"DUMMY", nil}})

type dummymodule struct {
	SimpleModule
}
