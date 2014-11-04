package module

import (
	"log"
)

/*
Defines a EyeStream local module.
All modules that are to interact with beenleigh and other local modules needs to implement this.
*/
type Module interface {
	Namer
	Logger() Logger
	New(map[string]interface{}) interface{} // Might be subject of change in future
}

//Any type capable of naming itself.
type Namer interface {
	Name() string
}

//Interface declaring what a module is allowed to do to a logger.
//More specifically, this is the "module-safe" subset of the common log methods,
//meaning they do not alter execution or cause panics.
//Modules are also not allowed to modify their Logger. This is done for them.
//Designed to be implemented by the *Logger type from the log package of the standard library.
type Logger interface {
	Print(...interface{})
	Printf(string, ...interface{})
	Println(...interface{})
}

//"Abstract" implementation of Module.
//Reduces boiler-plate code in most modules.
//If any feature is added to Module, it should first be attempted to be implemented here.
//This is to reduce the amount of code that needs to be changed.
type SimpleModule struct {
	Title   string
	Logbook Logger
}

func (sm SimpleModule) Name() string {
	return sm.Title
}

func (sm SimpleModule) Logger() Logger {
	return sm.Logbook
}
