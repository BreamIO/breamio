package hotkeys

import (
	"github.com/maxnordlund/breamio/beenleigh"
	//"github.com/maxnordlund/breamio/briee"

	"log"
	"os"
)

//#include "hotkey.h"
//#include <windows.h>
import "C"

const WM_HOTKEY = 0x0312

const (
	MOD_ALT = 1 << iota
	MOD_CONTROL
	MOD_SHIFT
	MOD_WIN
	MOD_NOREPEAT = 0x4000
)

var logger *log.Logger

func init() {
	beenleigh.Register(beenleigh.NewRunHandler(start))
	logger = log.New(os.Stderr, "[Hotkeys] ", log.LstdFlags)
}

func start(logic beenleigh.Logic, closer <-chan struct{}) {
	shutdown := false
	altBId := C.ES_RegisterHotKey(MOD_ALT|MOD_NOREPEAT, 0x42)
	if altBId == 0 {
		logger.Println("Unable to register hotkey ALT-B.")
	}

	go func() {
		<-closer
		shutdown = true
		logger.Println("Shutting down.")
		C.UnregisterHotKey(nil, altBId)
	}()

	var msg C.MSG

	drawing := true
	emitter := logic.CreateEmitter(1)
	resume := emitter.Publish("drawer:resume", struct{}{}).(chan<- struct{})
	pause := emitter.Publish("drawer:pause", struct{}{}).(chan<- struct{})

	defer close(resume)
	defer close(pause)

	for C.ES_GetMessage(&msg) != 0 || !shutdown {
		if msg.message == WM_HOTKEY {
			logger.Printf("Hotkey %s detected!", msg.lParam)
			if drawing {
				pause <- struct{}{}
			} else {
				resume <- struct{}{}
			}
			drawing = !drawing
		}
	}
}
