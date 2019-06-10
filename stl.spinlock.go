package stl

import (
	"runtime"
	"sync"
	"sync/atomic"
)

type SpinLock struct {
	_    sync.Mutex
	lock uintptr
}

func (l *SpinLock) Lock() {
	for !atomic.CompareAndSwapUintptr(&l.lock, 0, 1) {
		runtime.Gosched()
	}
}

func (l *SpinLock) Unlock() {
	atomic.StoreUintptr(&l.lock, 0)
}
