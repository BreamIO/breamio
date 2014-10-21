package main

import (
	"github.com/maxnordlund/breamio/aioli/client"
	"log"
	"os"
)

type Registrater interface {
	Register(Key) <-chan Key
	Unregister(Key)
}

var (
	r      Registrater
	logger *log.Logger
)

func init() {
	logger = log.New(os.Stderr, "[Hotkeys] ", log.LstdFlags)
}

func main() {
	shutdown := false
	key := Key{Alt, 'B'}

	presses := r.Register(key)
	defer r.Unregister(key)

	go func() {
		<-closer
		shutdown = true
		logger.Println("Shutting down.")
	}()

	var msg C.MSG
	drawing := true
	for C.ES_GetMessage(&msg) != 0 || !shutdown {
		if msg.message == WM_HOTKEY {
			if drawing {
				logic.CreateEmitter(1).Dispatch("drawer:pause", struct{}{})
				logger.Println("Pausing drawers.")
			} else {
				logic.CreateEmitter(1).Dispatch("drawer:resume", struct{}{})
				logger.Println("Resuming drawers.")
			}
			//Creating an emitter causes the TODO Continue comment

			drawing = !drawing
		}
	}

	logger.Println("This should not happen yet...")
}
