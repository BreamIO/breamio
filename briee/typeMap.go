package briee

import (
	"reflect"
	"sync"
)

var (
	typeMap     = make(map[string]reflect.Type)
	typeMapLock = sync.Locker(new(sync.Mutex))
)

func RegisterGlobalEventType(event string, typ reflect.Type) {
	defer typeMapLock.Unlock()
	typeMapLock.Lock()
	typeMap[event] = typ
}
