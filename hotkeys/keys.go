package main

const (
	Alt byte = 1 << iota
	Ctrl
	Shift
	Meta
)

//Represents a key combination
//The Modifiers is a bit-field, where each bit is ORed according to the above constants.
//The Key is represented as the ASCII value of the uppercase variant of the key to be pressed.
type Key struct {
	Modifiers, Key byte
}
