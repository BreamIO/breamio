package comte

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

const DefaultConfigFile = "settings.json"

/*
Represents any type that is configurable.
A configurable type needs to have a name and
a default layout as specified by the return type
of its Config method.
*/
type Configable interface {
	fmt.Stringer
	Config() ConfigSection
}

/*
Representation of a configuration.
Each section is identified with a key.
Once the configuration is loaded, the values can be retried using Section.
*/
type Configuration map[string]*json.RawMessage
type ConfigSection interface{}

//Default Configuration for use by most
var config = make(Configuration)

//Loads the configuration from the reader.
//JSON encoding is used.
func (c *Configuration) Load(in io.Reader) error {
	dec := json.NewDecoder(in)

	if err := dec.Decode(c); err != nil {
		return err
	}

	return nil
}

//Returns the section corresponding to the key.
//key matches return value of Configurables name.
//Returns nil if no section is defined for key or if nil value is stored.
func (c Configuration) Section(key string, mall ConfigSection) (cs ConfigSection) {
	if mall == nil {
		v := make(map[string]interface{})
		mall = &v
		defer func() {
			cs = v
		}()
	}
	json.Unmarshal(*c[key], mall)
	return mall
}

//Overwrites a stored configuration section with a new one.
//Section must already exist and the sections must be of the same type.
func (c Configuration) Update(key string, section ConfigSection) {
	jayson, err := json.Marshal(section)
	if err != nil {
		panic(err)
	}
	c[key] = (*json.RawMessage)(&jayson)
}

//Calls Load on the default configuration
func Load(in io.Reader) error {
	return config.Load(in)
}

//Calls Section on the default configuration
func Section(key string, cs ConfigSection) ConfigSection {
	return config.Section(key, cs)
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
