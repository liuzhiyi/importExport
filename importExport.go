package importExport

import (
	"github.com/liuzhiyi/importExport/storage"
)

var (
	imports = make(map[string]Import)
	exports = make(map[string]Export)
)

type Import interface {
	New(storage storage.Storage) Import
	Init(colNames ...string)
	Load(src string) error
	Next() error
	Key() int64
	GetColNames() []string
	Current() map[string]string
	Seek() error
	Close()
	To() error
}

type Export interface {
	New(storage storage.Storage) Export
	From() [][]string
	WriteRow(records []string) error
	Init(dst string, colNames ...string) error
	WriteAll() error
	Filter(where string)
	SetRegexp(reg map[string]string)
	Close()
}

func RegisterImport(name string, i Import) {
	imports[name] = i
}

func RegisterExport(name string, e Export) {
	exports[name] = e
}

func NewImport(name string, storage storage.Storage) Import {
	if i, ok := imports[name]; ok {
		return i.New(storage)
	} else {
		panic(name + " don't register")
	}
}

func NewExport(name string, storage storage.Storage) Export {
	if e, ok := exports[name]; ok {
		return e.New(storage)
	} else {
		panic(name + " don't register")
	}
}
