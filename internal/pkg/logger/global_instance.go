package logger

import "sync"

var (
	gl     LoggerV1
	lMutex sync.RWMutex
)

func SetGlobalLogger(l LoggerV1) {
	lMutex.Lock()
	defer lMutex.Unlock()
	gl = l
}

func L() LoggerV1 {
	lMutex.RLock()
	g := gl
	lMutex.Unlock()
	return g
}
