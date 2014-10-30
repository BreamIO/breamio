package module

import (
	"log"
	"os"
)

type Logging interface {
	Nameable
	Logger() *log.Logger
	SetLogger(*log.Logger)
}

func NewLogger(n Nameable) *log.Logger {
	return log.New(os.Stderr, "[ "+n.Name()+" ] ", log.LstdFlags|log.Lshortfile)
}
