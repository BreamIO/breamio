package main

import (
	"github.com/maxnordlund/breamio/beenleigh"
	//"github.com/maxnordlund/breamio/briee"

	"log"
	"os"
)

//#include "hotkey_windows.h"
//#include <windows.h>
import "C"

const WM_HOTKEY = 0x0312

type WindowsRegistrer map[Key]struct{id int, ch chan Key}

func (wr WindowsRegistrer) Register(k Key) {
	wr[k] = ES_RegisterHotKey(k.Modifiers, k.Key)
}


func (wr WindowsRegistrer) Unregister(k Key) {
	if v, ok := wr[k]; ok {
	}
}
