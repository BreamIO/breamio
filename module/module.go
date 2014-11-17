package module

import (
	"fmt"
)

/*
Defines a EyeStream local module.
All modules that are to interact with beenleigh and other local modules needs to implement this.
*/
type Module interface {
	fmt.Stringer
	Logger() Logger
}

type Factory interface {
	fmt.Stringer
	New(Constructor) Module // Might be subject of change in future
}

//Interface declaring what a module is allowed to do to a logger.
//Modules are also not allowed to modify their Logger. This is done for them.
//Designed to be implemented by the *Logger type from the log package of the standard library.
type Logger interface {
	Print(...interface{})
	Printf(string, ...interface{})
	Println(...interface{})

	Fatal(...interface{})
	Fatalf(string, ...interface{})
	Fatalln(...interface{})

	Panic(...interface{})
	Panicf(string, ...interface{})
	Panicln(...interface{})
}

type Constructor struct {
	Logger     Logger
	Parameters map[string]interface{}
}

//Type allowing struct-tagging of method
//to allow special instructions regarding exact event name and such
type EventMethod struct{}

//"Abstract" implementation of Module.
//Reduces boiler-plate code in most modules.
//If any feature is added to Module, it should first be attempted to be implemented here.
//This is to reduce the amount of code that needs to be changed.
type SimpleModule struct {
	name   string
	logger Logger
}

func NewSimpleModule(name string, c Constructor) SimpleModule {
	return SimpleModule{name, c.Logger}
}

func (sm SimpleModule) String() string {
	return sm.name
}

func (sm SimpleModule) Logger() Logger {
	return sm.logger
}
