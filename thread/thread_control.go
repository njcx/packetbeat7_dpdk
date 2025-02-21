package thread

import (
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
)

var (
	mainThreadID   int64
	mainThreadOnce sync.Once
	mainThreadChan = make(chan func() error, 1)
	isMainThread   uint32
)

func InitMainThread() {
	mainThreadOnce.Do(func() {
		runtime.LockOSThread()
		mainThreadID = getCurrentThreadID()
		atomic.StoreUint32(&isMainThread, 1)
		go mainThreadScheduler()
	})
}

func mainThreadScheduler() {
	for f := range mainThreadChan {
		f()
	}
}

func ExecuteInMainThread(f func() error) error {
	if atomic.LoadUint32(&isMainThread) == 0 {
		errChan := make(chan error, 1)
		mainThreadChan <- func() error {
			err := f()
			errChan <- err
			return err
		}
		return <-errChan
	}
	return f()
}

func IsMainThread() bool {
	return atomic.LoadUint32(&isMainThread) == 1
}

func getCurrentThreadID() int64 {
	return int64(syscall.Gettid())
}
