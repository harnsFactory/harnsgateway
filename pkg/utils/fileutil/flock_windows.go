package fileutil

import (
	"harnsgateway/pkg/tsdb/fileutil"
	"os"
	"syscall"
	"unsafe"
)

// https://docs.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-lockfileex
// https://docs.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-unlockfileex
var (
	modkernel32      = syscall.NewLazyDLL("kernel32.dll")
	procLockFileEx   = modkernel32.NewProc("LockFileEx")
	procUnlockFileEx = modkernel32.NewProc("UnlockFileEx")
)

const (
	_lockFileExclusiveLock   = 0x00000002
	_lockFileFailImmediately = 0x00000001
)

type windowsLock struct {
	fd syscall.Handle
}

var _ fileutil.Releaser = (*windowsLock)(nil)

func (fl *windowsLock) Release() error {
	return unlockFileEx(fl.fd, 1, 0, &syscall.Overlapped{})
}

// lookup err from https://docs.microsoft.com/zh-cn/windows/win32/debug/system-error-codes--0-499-?redirectedfrom=MSDN
func (fl *windowsLock) lock() error {
	return lockFileEx(fl.fd, _lockFileExclusiveLock|_lockFileFailImmediately, 1, 0, &syscall.Overlapped{})
}

func NewLock(f *os.File) (fileutil.Releaser, error) {
	l := &windowsLock{syscall.Handle(f.Fd())}
	return l, l.lock()
}

func lockFileEx(h syscall.Handle, flags, locklow, lockhigh uint32, ol *syscall.Overlapped) (err error) {
	var reserved uint32 = 0
	r1, _, e1 := syscall.Syscall6(procLockFileEx.Addr(), 6, uintptr(h), uintptr(flags), uintptr(reserved), uintptr(locklow), uintptr(lockhigh), uintptr(unsafe.Pointer(ol)))
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return err
}

func unlockFileEx(h syscall.Handle, locklow, lockhigh uint32, ol *syscall.Overlapped) (err error) {
	var reserved uint32 = 0
	r1, _, e1 := syscall.Syscall6(procUnlockFileEx.Addr(), 5, uintptr(h), uintptr(reserved), uintptr(locklow), uintptr(lockhigh), uintptr(unsafe.Pointer(ol)), 0)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return err
}
