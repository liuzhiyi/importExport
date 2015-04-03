package importExport

import (
	"encoding/csv"
	"errors"
	"fmt"
	"os"

	"github.com/liuzhiyi/importExport/storage"
)

var (
	isSetColsErr       = errors.New("Header column names already set")
	emptyColsErr       = errors.New("Header column names don't set")
	defaultSize  int64 = 100
)

func init() {
	RegisterExport("csv", &csvExport{})
	RegisterImport("csv", &csvImort{})
}

type csvImort struct {
	colNames   []string
	currentRow []string
	currentKey int64
	source     string
	storage    storage.Storage
	size       int64
	r          *csv.Reader
	f          *os.File
}

func (c *csvImort) New(storage storage.Storage) Import {
	return &csvImort{
		storage: storage,
	}
}

func (c *csvImort) Init(colNames ...string) {
	c.colNames = colNames
}

func (c *csvImort) Load(src string) error {
	f, err := os.OpenFile(src, os.O_RDONLY, 0)
	if err != nil {
		fmt.Print(err.Error())
		return err
	}

	c.source = src
	c.r = csv.NewReader(f)
	c.f = f
	//err = c.rewind()

	return err
}

func (c *csvImort) GetColNames() []string {
	return c.colNames
}

func (c *csvImort) rewind() error {
	var err error

	c.f.Seek(0, 0)
	c.colNames, err = c.r.Read()
	if err != nil {
		return err
	}

	err = c.Next()
	if err != nil {
		return err
	}
	c.currentKey = 0

	return nil
}

func (c *csvImort) Current() map[string]string {
	row := make(map[string]string)
	for i, col := range c.colNames {
		row[col] = c.currentRow[i]
	}
	return row
}

func (c *csvImort) Key() int64 {
	return c.currentKey
}

func (c *csvImort) Seek() error {
	return nil
}

func (c *csvImort) Next() error {
	records, err := c.r.Read()
	if err != nil {
		return err
	}

	c.currentRow = records
	c.currentKey += 1

	return nil
}

func (c *csvImort) To() error {
	c.storage.SetConfig("fileName", c.source)
	c.storage.RawWrite(c.f)
	return nil
}

func (c *csvImort) Close() {
	if c.f != nil {
		c.f.Close()
	}
}

type csvExport struct {
	destination string
	colNames    []string
	over        bool
	isSet       bool
	size        int64
	storage     storage.Storage
	f           *os.File
	w           *csv.Writer
}

func (c *csvExport) New(storage storage.Storage) Export {
	return &csvExport{
		storage: storage,
	}
}

func (c *csvExport) Init(dst string, colNames ...string) error {
	f, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0)
	if err != nil {
		return err
	}

	c.f = f
	c.size = defaultSize
	c.w = csv.NewWriter(f)

	err = c.setHeaderCols(colNames)

	if err != nil {
		return err
	}
	return err
}

func (c *csvExport) Filter(where string) {
	c.storage.Filter(where)
}

func (c *csvExport) From() [][]string {
	var records [][]string

	c.storage.SetSize(c.size)
	records = c.storage.Read()

	if len(records) == 0 {
		c.over = true
	}
	return records
}

func (c *csvExport) WriteRow(records []string) error {
	if !c.isSet {
		return emptyColsErr
	}
	return c.w.Write(records)
}

func (c *csvExport) GetColNames() []string {
	return c.colNames
}

func (c *csvExport) WriteAll() error {
	var (
		err     error
		records [][]string
	)

	for !c.over && err == nil {
		records = c.From()
		for _, record := range records {
			err = c.WriteRow(record)
		}
	}
	return err
}

func (c *csvExport) setHeaderCols(cols []string) error {
	if c.isSet {
		return isSetColsErr
	}
	c.colNames = cols
	c.storage.SetColNames(cols)
	c.isSet = true
	return c.w.Write(cols)
}

func (c *csvExport) Close() {
	c.w.Flush()
	if c.f != nil {
		c.f.Close()
	}
}

func (c *csvExport) SetRegexp(reg map[string]string) {
	c.storage.SetRegexp(reg)
}
