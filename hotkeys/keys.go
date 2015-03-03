package main

import (
	"fmt"
	"strings"
)

//Represents a key combination
//The Modifiers is a bit-field, where each bit is ORed according to the above constants.
//The Key is represented as the ASCII value of the uppercase variant of the key to be pressed.
type Key struct {
	Modifiers Modifier
	Key       byte
}

func (k Key) String() string {
	return fmt.Sprintf("%s-%c", k.Modifiers, k.Key)
}

type Modifier byte

const (
	Alt Modifier = 1 << iota
	Ctrl
	Shift
	Meta
)

func (m Modifier) String() string {
	parts := make([]string, 0, 4)
	if m&Ctrl != 0 {
		parts = append(parts, "ctrl")
	}
	if m&Shift != 0 {
		parts = append(parts, "shift")
	}
	if m&Alt != 0 {
		parts = append(parts, "alt")
	}

	if m&Meta != 0 {
		parts = append(parts, "win")
	}

	return strings.Join(parts, "-")
}
