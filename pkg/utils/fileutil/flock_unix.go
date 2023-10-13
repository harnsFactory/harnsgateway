//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd
// +build darwin dragonfly freebsd linux netbsd openbsd

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
	how := syscall.LOCK_UN
	if lock {
		how = syscall.LOCK_EX
	}
	return syscall.Flock(int(l.f.Fd()), how|syscall.LOCK_NB)
}

func NewLock(f *os.File) (fileutil.Releaser, error) {
	l := &unixLock{f}
	return l, l.set(true)
}
