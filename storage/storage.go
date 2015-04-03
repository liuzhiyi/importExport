package storage

import "io"

var storages = make(map[string]Storage)

func Register(name string, storage Storage) {
	if storage == nil {
		panic("package storage: register a storage is nil")
	}
	storages[name] = storage
}

func New(name string) Storage {
	if storage, ok := storages[name]; ok {
		return storage.New()
	} else {
		panic("storage: " + name + " storage is not exist")
	}
}

type Storage interface {
	New() Storage
	Init(config map[string]string)
	SetColNames(cols []string)
	Read() [][]string
	RawWrite(reader io.Reader)
	SetSize(size int64)
	SetCursor(cursor int64)
	Filter(where string)
	SetRegexp(reg map[string]string)
	SetConfig(key, val string) error
}

type StorageParam interface {
	SetParam(data string)
	GetParam() string
}
