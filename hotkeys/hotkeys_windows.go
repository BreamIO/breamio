package main

import (
	"sync"
)

//#include "hotkey_windows.h"
//#include <windows.h>
import "C"

const WM_HOTKEY = 0x0312
const NO_REPEAT = 0x4000
const word = 16

func init() {
	wr := &WindowsRegistrer{
		bindings: make(map[Key]binding),
		Locker:   &sync.Mutex{},
		closed:   make(chan error),
	}
	go wr.run()
	r = wr
}

type WindowsRegistrer struct {
	bindings map[Key]binding
	shutdown bool
	closed   chan error
	sync.Locker
}

type binding struct {
	id int
	ch chan<- Key
}

func (wr WindowsRegistrer) Register(k Key, ch chan<- Key) error {
	wr.Lock()
	defer wr.Unlock()

	if _, ok := wr.bindings[k]; ok {
		return DoubleBinding
	}

	logger.Println("Registering key:", k)
	id := C.ES_RegisterHotKey(C.uint(k.Modifiers)|NO_REPEAT, C.uint(k.Key))
	wr.bindings[k] = binding{int(id), ch}
	return nil
}

func (wr WindowsRegistrer) Unregister(k Key) {
	wr.Lock()
	defer wr.Unlock()

	if v, ok := wr.bindings[k]; ok {
		logger.Println("Unregistering key:", k)
		C.UnregisterHotKey(nil, C.int(v.id))
		delete(wr.bindings, k)
	}
}

func (wr *WindowsRegistrer) Close() error {
	wr.shutdown = true
	return <-wr.closed
}

func (wr *WindowsRegistrer) run() {
	defer close(wr.closed)

	for {
		var msg C.MSG
		if C.ES_GetMessage(&msg) != 0 {
			// logger.Println("Message Get!")
			if msg.message == WM_HOTKEY {
				key := Key{Modifier(msg.lParam), byte(msg.lParam >> word)}
				wr.handle(key)
			}
		}
		if wr.shutdown {
			return
		}
	}
}

func (wr WindowsRegistrer) handle(key Key) {
	logger.Println("Key:", key)
	wr.Lock()
	defer wr.Unlock()
	if val, ok := wr.bindings[key]; ok {
		val.ch <- key
	}

}
