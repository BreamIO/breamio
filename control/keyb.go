// +build windows

package control

/*
#include <windows.h>
#include <winable.h>

int keyboardInput(INPUT *in, unsigned int key, boolean down) {
	in->type = INPUT_KEYBOARD;
	in->ki.wVk = (WORD) key;
	if (!down) {
		in->ki.dwFlags = KEYEVENTF_KEYUP;
	}
	return sizeof(*in);
}

*/
import "C"

import (
	region "github.com/maxnordlund/breamio/analysis/regionStats"
	bl "github.com/maxnordlund/breamio/beenleigh"
	"github.com/maxnordlund/breamio/briee"
	"github.com/maxnordlund/breamio/gorgonzola"
	"log"
	"os"
	"syscall"
	"unsafe"
)

const (
	n = 1
	margin = 0
	width = 0.3
)

var logger *log.Logger
var sendInput *syscall.Proc
var areas map[region.Region]func(in bool)

func init() {
	logger = log.New(os.Stdout, "[control] ", log.LstdFlags)

	dll, err := syscall.LoadDLL("user32")
	if err != nil {
		logger.Println("Could not load user32.dll: ", err)
		return
	}

	sendInput, err = dll.FindProc("SendInput")
	if err != nil {
		logger.Println("Could not find procedure SendInput: ", err)
		return
	}

	areas = make(map[region.Region]func(bool))
	up, _ := region.NewRegion("up", region.RegionDefinition{"rect", margin, 1.0-2*margin, margin, width})
	areas[up] = keybPresser(C.VK_UP)
	
	left, _ := region.NewRegion("left", region.RegionDefinition{"rect", margin, width, margin, 1.0-2*margin})
	areas[left] = keybPresser(C.VK_LEFT)

	down, _ := region.NewRegion("down", region.RegionDefinition{"rect", margin, 1.0-2*margin, 1.0-margin-width, width})
	areas[down] = keybPresser(C.VK_DOWN)
	
	right, _ := region.NewRegion("right", region.RegionDefinition{"rect", 1.0-margin-width, width, margin, 1.0-2*margin})
	areas[right] = keybPresser(C.VK_RIGHT)

	bl.Register(bl.NewRunHandler(func(logic bl.Logic, closer <-chan struct{}) {
		logger.Println("Initializing Control subsystem")
		newCh := logic.RootEmitter().Subscribe("new:control", bl.Spec{}).(<-chan bl.Spec)
		defer logic.RootEmitter().Unsubscribe("new:control", newCh)
		
		for {
			select {
				case spec, ok := <-newCh: 
					if ok {
						logger.Printf("New controller on emitter %d\n", spec.Emitter)
						go handle(logic.CreateEmitter(spec.Emitter))
					}
				case <-closer: return
			}
		}
	}))
}

func handle(ee briee.EventEmitter) {
	etCh := ee.Subscribe("tracker:etdata", &gorgonzola.ETData{}).(<-chan *gorgonzola.ETData)
	defer ee.Unsubscribe("tracker:etdata", etCh)
	
	//shutdownCh := ee.Subscribe("control:shutdown", struct{}{})

	for data := range etCh {
		for region, f := range areas {
			f(region.Contains(data.Filtered))
		}
	}
}

func keybPresser(button uint) func(bool) {
	down := false
	return func(in bool) {
		event := &C.INPUT{}
		if in && !down {
			//logger.Printf("Pressing %x.", button)
			size := C.keyboardInput(event, C.uint(button), CBool(true)) //press
			//log.Println(in, down, event)
			sendInput.Call(uintptr(n), uintptr(unsafe.Pointer(event)), uintptr(size))
			down = true
		} else if !in && down {
			size := C.keyboardInput(event, C.uint(button), CBool(false)) // release
			//log.Println(event)
			sendInput.Call(uintptr(n), uintptr(unsafe.Pointer(event)), uintptr(size))
			down = false
		}
	}
}

func CBool(b bool) C.boolean {
	if b {
		return C.boolean(1)
	}
	return C.boolean(0)
}
