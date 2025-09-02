package registry

import "sync"

type Registry struct {
	mu   sync.RWMutex
	objs map[string]interface{}
}

var global = &Registry{objs: make(map[string]interface{})}

func Set(name string, obj interface{}) {
	global.mu.Lock()
	defer global.mu.Unlock()
	global.objs[name] = obj
}

func Get(name string) (interface{}, bool) {
	global.mu.RLock()
	defer global.mu.RUnlock()
	v, ok := global.objs[name]
	return v, ok
}

func MustGet(name string) interface{} {
	if v, ok := Get(name); ok {
		return v
	}
	panic("registry: object not found: " + name)
}
