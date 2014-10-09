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

	C.UnregisterHotKey(nil, 1) //Clear an existing registration
	altBId := C.ES_RegisterHotKey(MOD_ALT|MOD_NOREPEAT, 0x42)
	defer C.UnregisterHotKey(nil, altBId)

	if altBId == 0 {
		logger.Println("Unable to register hotkey ALT-B.")
	}

	logger.Println("Listening for ALT-B to toggle drawing.")

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
