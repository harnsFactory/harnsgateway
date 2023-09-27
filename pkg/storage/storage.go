package storage

import (
	"time"
)

type StoreGroup byte

const (
	StoreGroupDevice StoreGroup = iota
)

var (
	StoreGroupToString = map[StoreGroup]string{
		StoreGroupDevice: "device",
	}
	StoreGroupFromString = map[string]StoreGroup{
		"device": StoreGroupDevice,
	}
)

// resources
const (
	// device
	Devices = "devices"
)

type Getter interface {
	Get(key string) (interface{}, error)
}

type Lister interface {
	List(key string) (interface{}, error)
}

type Creater interface {
	Create(key string, obj interface{}) (interface{}, error)
}

type Updater interface {
	Update(key, version string, obj interface{}) (interface{}, error)
}

type Deleter interface {
	Delete(key, version string) (interface{}, error)
}

// type Watcher interface {
// 	Watch(stopCh <-chan struct{}, key string, rev string) (chan *Event, error)
// }

type Storage interface {
	Getter
	Lister
	Creater
	Updater
	Deleter
	// Watcher
}

type EventType int8

const (
	Create EventType = iota
	Update
	Remove
)

func (et EventType) String() string {
	return []string{
		"Create",
		"Update",
		"Remove",
	}[et]
}

type Event struct {
	Type EventType
	Data interface{}
}

type FileInfo struct {
	Path    string
	ModTime time.Time
}
