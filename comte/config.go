package comte

import (
	"encoding/json"
	"errors"
	"io"
	"reflect"
)

const DefaultConfigFile = "config.json"

/*
Represents any type that is configurable.
A configurable type needs to have a name and
a default layout as specified by the return type
of its Config method.
*/
type Configable interface {
	Name() string
	Config() ConfigSection
}

/*
Representation of a configuration.
Each section is identified with a key.
Once the configuration is loaded, the values can be retried using Section.
*/
type Configuration map[string]ConfigSection
type ConfigSection interface{}

//Default Configuration for use by most
var config = make(Configuration)

//Registers a configurable for use in this Configuration.
func (c Configuration) Register(module Configable) {
	config[module.Name()] = module.Config()
}

//Loads the configuration from the reader.
//JSON encoding is used.
func (c Configuration) Load(in io.Reader) error {
	dec := json.NewDecoder(in)
	if dec.Decode(c) != nil {
		return Undecodable
	}
	return nil
}

//Returns the section corresponding to the key.
//key matches return value of Configurables name.
//Returns nil if no section is defined for key or if nil value is stored.
func (c Configuration) Section(key string) ConfigSection {
	return c[key]
}

//Overwrites a stored configuration section with a new one.
//Section must already exist and the sections must be of the same type.
func (c Configuration) Update(key string, section ConfigSection) {
	if _, ok := c[key]; !ok {
		panic(NonexistingKey)
	}

	if reflect.TypeOf(section) != reflect.TypeOf(c[key]) {
		panic(BadSection)
	}
}

//Calls Register on the default configuration
func Register(module Configable) {
	config.Register(module)
}

//Calls Load on the default configuration
func Load(in io.Reader) error {
	return config.Load(in)
}

//Calls Section on the default configuration
func Section(key string) ConfigSection {
	return config.Section(key)
}

//Calls Update on the default configuration
func Update(key string, section ConfigSection) {
	config.Update(key, section)
}

var (
	//Error signaling a undecodable input
	Undecodable = errors.New("Undecodable input")

	//Error signaling a key not existing in the configuration
	NonexistingKey = errors.New("Nonexisting key")

	//Error signaling a error with a passed section
	BadSection = errors.New("Bad Section")
)
