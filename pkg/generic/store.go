package generic

import (
	"encoding/json"
	"harnsgateway/pkg/runtime"
	"harnsgateway/pkg/storage"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

type Store struct {
	Group        string
	Resource     string
	ResourceType map[string]reflect.Type
	client       storage.Storage
}

func NewStore(group string, resource string, resourceType map[string]runtime.RunObject) (*Store, error) {
	s := &Store{
		Group:        group,
		Resource:     resource,
		ResourceType: make(map[string]reflect.Type),
	}
	for dt, object := range resourceType {
		s.ResourceType[dt] = getTypeOfResource(object)
	}

	client := &storage.FsClient{}
	client.Init(storage.StoreGroupFromString[group])
	s.client = client

	return s, nil
}

func (s *Store) Create(obj runtime.RunObject) (save runtime.RunObject, returnErr error) {
	accessor, _ := runtime.Accessor(obj)
	key := filepath.Join(s.Resource, accessor.GetID())
	if saved, err := s.client.Create(key, obj); err == nil {
		save = saved.(runtime.RunObject)
	} else {
		returnErr = err
	}
	return
}

func (s *Store) Update(obj runtime.RunObject) (update runtime.RunObject, returnErr error) {
	accessor, _ := runtime.Accessor(obj)
	key := filepath.Join(s.Resource, accessor.GetID())
	if updated, err := s.client.Update(key, accessor.GetVersion(), obj); err == nil {
		update = updated.(runtime.RunObject)
	} else {
		returnErr = err
	}
	return
}

func (s *Store) Delete(obj runtime.RunObject) (delete runtime.RunObject, returnErr error) {
	accessor, _ := runtime.Accessor(obj)
	key := filepath.Join(s.Resource, accessor.GetID())
	if _, err := s.client.Delete(key, accessor.GetVersion()); err == nil {
		delete = obj
	} else {
		returnErr = err
	}
	return
}

func (s *Store) LoadResource() ([]runtime.RunObject, error) {
	objs, err := s.client.List(s.Resource)
	if err != nil {
		return nil, err
	}

	var ret []runtime.RunObject
	if files, ok := objs.([]*storage.FileInfo); ok {
		for _, file := range files {
			func() {
				fileName := filepath.Base(file.Path)
				dt := fileName[0:strings.LastIndex(fileName, ".")]
				obj := reflect.New(s.ResourceType[dt]).Interface().(runtime.RunObject)
				f, err := os.Open(file.Path)
				defer f.Close()
				if err != nil {
					klog.V(2).InfoS("Failed to open", "file", file.Path, "resource", s.Resource, "err", err)
					return
				}
				if err = json.NewDecoder(f).Decode(obj); err != nil {
					klog.V(3).InfoS("Failed to unmarshal", "file", file.Path, "resource", s.Resource, "err", err)
					return
				}
				ret = append(ret, obj)
			}()
		}
	}
	return ret, nil
}

func getTypeOfResource(obj runtime.RunObject) reflect.Type {
	t := reflect.TypeOf(obj)
	if t.Kind() != reflect.Ptr {
		panic("All types must be pointers to structs.")
	}
	t = t.Elem()
	if t.Kind() != reflect.Struct {
		panic("All types must be pointers to structs.")
	}
	return t
}
