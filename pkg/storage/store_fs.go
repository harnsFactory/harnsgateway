package storage

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"golang.org/x/mod/sumdb"
	"harnsgateway/pkg/apis"
	"harnsgateway/pkg/runtime"
	"harnsgateway/pkg/utils/fileutil"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

type FsClient struct {
	storePath string
}

var (
	_ Storage = (*FsClient)(nil)

	doOnce sync.Once
)

func (fc *FsClient) Init(sg StoreGroup) {
	_, err := os.Stat(storePath)
	if err != nil {
		klog.Fatalf("%s: %v", storePath, err)
	}

	var dirs []string
	switch sg {
	case StoreGroupDevice:
		dirs = []string{
			Devices,
		}
	default:
		klog.Fatalf("Unsupported store group %d", sg)
	}

	fc.storePath = filepath.Join(storePath, StoreGroupToString[sg])

	for _, m := range dirs {
		p := filepath.Join(fc.storePath, m)

		_, err = os.Stat(p)
		if os.IsNotExist(err) {
			absPath, _ := filepath.Abs(p)
			klog.V(2).InfoS("Created", "path", absPath)
			if err = os.MkdirAll(p, 0711); err != nil {
				klog.Fatal(err)
			}
		} else if err != nil {
			klog.Fatal(err)
		}
	}

	doOnce.Do(func() {
		gob.Register(map[string]interface{}{})
		gob.Register([]interface{}{})
	})
}

func (fc *FsClient) Create(key string, obj interface{}) (interface{}, error) {
	f, err := os.OpenFile(filepath.Join(fc.storePath, key), os.O_CREATE|os.O_RDWR|os.O_EXCL, 0640)
	if err != nil {
		klog.V(2).InfoS("Failed to create file", "err", err)
		return nil, err
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(obj)
	if err != nil {
		klog.V(2).InfoS("Failed to encode", "err", err)
		return nil, err
	}
	return obj, nil
}

func (fc *FsClient) Get(key string) (interface{}, error) {
	data, err := os.ReadFile(filepath.Join(fc.storePath, key))
	if err != nil {
		klog.V(2).InfoS("Failed to read", "err", err)
		return nil, err
	}
	return data, nil
}

func (fc *FsClient) List(key string) (interface{}, error) {
	var files []*FileInfo
	err := filepath.Walk(filepath.Join(fc.storePath, key), func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, &FileInfo{
				Path: path,
			})
		}
		return nil
	})
	if err != nil {
		klog.V(2).InfoS("Failed to list", "err", err)
	}
	return files, nil
}

func (fc *FsClient) Delete(key, version string) (interface{}, error) {
	// version is not required when cascading delete
	if len(version) == 0 {
		c, cancel := context.WithCancel(context.Background())
		wait.UntilWithContext(c, func(ctx context.Context) {
			if err := os.Remove(filepath.Join(fc.storePath, key)); !isEphemeralError(err) {
				if err != nil {
					klog.V(5).InfoS("Failed to remove file", "err", err)
				}
				cancel()
			}
		}, 0)
		return nil, nil
	}

	f, err := os.OpenFile(filepath.Join(fc.storePath, key), os.O_RDONLY, 0640)
	if err != nil {
		if os.IsNotExist(err) {
			klog.V(2).InfoS("Failed to open file", "err", err)
			return nil, os.ErrNotExist
		} else if isEphemeralError(err) {
			klog.V(2).InfoS("Failed to open file", "err", err)
			return nil, sumdb.ErrWriteConflict
		}
	}
	defer f.Close()

	lock, err := fileutil.NewLock(f)
	if err != nil {
		klog.V(2).InfoS("Failed to lock", "err", err)
		return nil, sumdb.ErrWriteConflict
	}
	defer lock.Release()

	var target struct {
		runtime.ObjectMeta
	}
	err = json.NewDecoder(f).Decode(&target)
	if err != nil {
		klog.V(2).InfoS("Failed to unmarshal", "err", err)
		return nil, apis.ErrInternal
	}
	if target.Version != version {
		return nil, apis.ErrMismatch
	}
	_ = f.Close()

	err = os.Remove(filepath.Join(fc.storePath, key))
	if err != nil {
		klog.V(2).InfoS("Failed to remove", "err", err)
		return nil, apis.ErrInternal
	}
	return nil, nil
}

func (fc *FsClient) Update(key, version string, obj interface{}) (interface{}, error) {
	f, err := os.OpenFile(filepath.Join(fc.storePath, key), os.O_RDWR, 0640)
	if err != nil {
		if os.IsNotExist(err) {
			klog.V(2).InfoS("Failed to open file", "err", err)
			return nil, os.ErrNotExist
		} else if isEphemeralError(err) {
			klog.V(2).InfoS("Failed to open file", "err", err)
			return nil, sumdb.ErrWriteConflict
		}
	}
	defer f.Close()

	lock, err := fileutil.NewLock(f)
	if err != nil {
		klog.V(2).InfoS("Failed to lock", "err", err)
		return nil, sumdb.ErrWriteConflict
	}
	defer lock.Release()

	var old struct {
		runtime.ObjectMeta
	}
	err = json.NewDecoder(f).Decode(&old)
	if err != nil {
		klog.V(2).InfoS("Failed to unmarshal", "err", err)
		return nil, apis.ErrInternal
	}
	if version != old.Version {
		return nil, apis.ErrMismatch
	}
	ver, _ := strconv.ParseUint(version, 10, 64)
	accessor, err := runtime.Accessor(obj)
	if err != nil {
		klog.V(2).InfoS("Failed to get accessor", "err", err)
		return nil, apis.ErrInternal
	}
	accessor.SetVersion(strconv.FormatUint(ver+uint64(rand.Intn(100)), 10))

	if err = f.Truncate(0); err != nil {
		klog.V(2).InfoS("Failed to truncate", "err", err)
		return nil, apis.ErrInternal
	}
	if _, err = f.Seek(0, 0); err != nil {
		klog.V(2).InfoS("Failed to seek", "err", err)
		return nil, apis.ErrInternal
	}
	err = json.NewEncoder(f).Encode(obj)
	if err != nil {
		klog.V(2).InfoS("Failed to marshal", "err", err)
		return nil, apis.ErrInternal
	}

	return obj, nil
}
