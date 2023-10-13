//go:build solaris
// +build solaris

package fileutil

import (
	"harnsgateway/pkg/tsdb/fileutil"
	"os"
	"syscall"
)

type unixLock struct {
	f *os.File
}

var _ fileutil.Releaser = (*unixLock)(nil)

func (l *unixLock) Release() error {
	return l.set(false)
}

func (l *unixLock) set(lock bool) error {
	flock := syscall.Flock_t{
		Type:   syscall.F_UNLCK,
		Start:  0,
		Len:    0,
		Whence: 1,
	}
	if lock {
		flock.Type = syscall.F_WRLCK
	}
	return syscall.FcntlFlock(l.f.Fd(), syscall.F_SETLK, &flock)
}

func NewLock(f *os.File) (fileutil.Releaser, error) {
	l := &unixLock{f}
	return l, l.set(true)
}
