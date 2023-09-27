package storage

import (
	"errors"
	"golang.org/x/sys/windows"
	"k8s.io/klog/v2"
	"os/user"
	"path/filepath"
	"syscall"
)

var (
	storePath = getStorePath()
)

func getStorePath() string {
	if u, err := user.Current(); err == nil {
		return filepath.Join(u.HomeDir, "harnsgateway")
	} else {
		klog.ErrorS(err, "Failed to get home dir")
		return "./harnsgateway"
	}
}

func isEphemeralError(err error) bool {
	var errno syscall.Errno
	if errors.As(err, &errno) {
		switch errno {
		case windows.ERROR_SHARING_VIOLATION:
			return true
		}
	}
	return false
}
