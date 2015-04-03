package storage

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	db "github.com/liuzhiyi/go-db"
)

func init() {
	Register("mysql", &mysql{})
}

var (
	dsnNoErr = errors.New("dsn: nonexistent key")
)

type mysql struct {
	colNames []string
	data     [][]string
	config   map[string]string
	regexp   map[string]string
	cursor   int64
	offset   int64
	count    int64
	db.Collection
}

func (m *mysql) New() Storage {
	return &mysql{}
}

func (m *mysql) Init(config map[string]string) {
	m.cursor = 1
	m.offset = 0

	m.initDsn()
	for key, val := range config {
		m.SetConfig(key, val)
	}
	readDsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?strict=false",
		m.config["username"],
		m.config["password"],
		m.config["host"],
		m.config["dbname"],
	)
	db.F.InitDb("mysql", readDsn, "")
}

func (m *mysql) initDsn() {
	m.config = make(map[string]string)
	m.config["username"] = "root"
	m.config["password"] = ""
	m.config["dbname"] = ""
	m.config["host"] = "localhost"
	m.config["charset"] = "utf-8"
	m.config["type"] = "mysql"
	m.config["tableName"] = ""
}

func (m *mysql) SetConfig(key, val string) error {
	if _, ok := m.config[key]; !ok {
		return dsnNoErr
	}
	m.config[key] = val
	if key == "tableName" {
		m.SetMainTable(val)
		m.Collection.Init(val)
	}
	return nil
}

func (m *mysql) SetColNames(cols []string) {
	m.colNames = cols
}

func (m *mysql) LoadCols() {
	fields := strings.Join(m.colNames, ", ")
	m.AddFieldToSelect(fields, m.GetMainAlias())
}

func (m *mysql) Filter(where string) {
	m.Collection.GetSelect().Where(where)
}

func (m *mysql) SetRegexp(reg map[string]string) {
	m.regexp = reg
}

func (m *mysql) processRecords(records []string) []string {
	for i, col := range m.colNames {
		reg, replace := m.getColReg(col)
		if reg != "" {
			regObj := regexp.MustCompile(reg)
			records[i] = regObj.ReplaceAllString(records[i], replace)
		}
	}
	return records
}

func (m *mysql) getColReg(name string) (reg, replace string) {
	if item, ok := m.regexp[name]; ok {
		item = strings.Trim(item, ",")
		i := strings.Index(item, ",")
		if i > 0 {
			reg = item[:i]
			replace = item[i+1:]
		} else {
			reg = item
			replace = ""
		}
		return
	}
	return "", ""
}

func (m *mysql) SetSize(size int64) {
	m.Collection.SetPageSize(size)
}

func (m *mysql) SetCursor(cursor int64) {
	m.cursor = cursor
}

func (m *mysql) Read() [][]string {
	var data [][]string

	if m.cursor > m.GetLastPage() {
		return data
	}

	m.LoadCols()
	m.Collection.SetCurPage(m.cursor)
	m.Load()
	m.cursor++

	for _, item := range m.GetItems() {
		records := m.processRecords(item.ToArray())
		data = append(data, records)
	}
	m.data = data

	return m.data
}

func (m *mysql) WriteRow() {

}

func (m *mysql) WriteAll() {

}

func (m *mysql) RawWrite(reader io.Reader) {

}
