package fileutil

import (
	"lightiot/pkg/tsdb/fileutil"
	"os"
)

type plan9Lock struct {
	f *os.File
}

var _ fileutil.Releaser = (*plan9Lock)(nil)

func (l *plan9Lock) Release() error {
	// return l.f.Close()
	panic("unsupported unlock file")
}

func NewLock(fileName string) (fileutil.Releaser, error) {
	// f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, os.ModeExclusive|0666)
	// if err != nil {
	//	return nil, err
	// }
	// return &plan9Lock{f}, nil
	panic("unsupported lock file")
}
