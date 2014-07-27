package briee

import (
	"reflect"
	"sync"
)

var (
	typeMap     map[string]reflect.Type
	typeMapLock sync.Locker
)

func RegisterGlobalEventType(event string, typ reflect.Type) {
	defer typeMapLock.Unlock()
	typeMapLock.Lock()
	typeMap[event] = typ
}
