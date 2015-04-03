package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/liuzhiyi/dockerAdmin/pkg/xml"
	"github.com/liuzhiyi/importExport"
	"github.com/liuzhiyi/importExport/storage"
)

const (
	col_xml_path       = "webList/%s/colNames"
	tablename_xml_path = "webList/%s/tablename"
	setup_xml_path     = "setup/%s"
	file_xml_path      = "webList/%s/file"
	filter_xml_path    = "webList/%s/filter"
	web_xml_path       = "webList"
)

func main() {
	logWrite, err := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY, 0)
	if err != nil {
		panic(err.Error())
	}
	logger := log.New(logWrite, "tool:", log.Llongfile|log.Ltime)
	config := xml.New()
	err = config.LoadFile("config.xml", nil)
	if err != nil {
		logger.Fatal(err.Error())
	}

	storageFrom := storage.New(config.SelectNode("", "storageFrom").GetValue())
	storageTo := storage.New(config.SelectNode("", "storageTo").GetValue())
	im := importExport.NewImport(config.SelectNode("", "fileType").GetValue(), storageTo)
	ex := importExport.NewExport(config.SelectNode("", "fileType").GetValue(), storageFrom)

	//init storage driver
	fromConf := getSetup(config, "storageFrom")
	toConf := getSetup(config, "storageTo")
	storageFrom.Init(fromConf)
	storageTo.Init(toConf)

	webNodes := config.SelectNodes("", web_xml_path)
	for _, webNode := range webNodes {
		webName := webNode.Name.Local

		//init importExport
		colNames, regexp := getColNames(config, webName)
		fileName := getNodeByPath(config.Root, fmt.Sprintf(file_xml_path, webName)).GetValue()
		err = ex.Init(fileName, colNames...)
		if err != nil {
			logger.Fatalf("export init:", err.Error())
		}
		im.Init(colNames...)
		where := getNodeByPath(config.Root, fmt.Sprintf(filter_xml_path, webName)).GetValue()
		ex.Filter(where)
		ex.SetRegexp(regexp)

		//release resource
		defer im.Close()
		defer ex.Close()

		//export csv from storage
		tableName := getNodeByPath(config.Root, fmt.Sprintf(tablename_xml_path, webName)).GetValue()
		storageFrom.SetConfig("tableName", tableName)
		if err = ex.WriteAll(); err != nil {
			logger.Printf("export write:%s", err.Error())
		}

		//import csv to storage
		err = im.Load(fileName)
		if err != nil {
			logger.Printf("import load file failure:%s", err.Error())
		}
		if err = im.To(); err != nil {
			logger.Printf("import to storage:%s", err.Error())
		}
	}
}

func getColNames(config *xml.Document, web string) ([]string, map[string]string) {
	colPath := fmt.Sprintf(col_xml_path, web)

	nodes := getNodeByPath(config.Root, colPath).Children
	colNames := []string{}
	regexp := make(map[string]string)
	for _, node := range nodes {
		if strings.Trim(node.Name.Local, " ") != "" {
			colNames = append(colNames, node.Name.Local)
			regexp[node.Name.Local] = node.GetValue()
		}
	}

	return colNames, regexp
}

func getSetup(config *xml.Document, storageType string) map[string]string {
	nodes := getNodeByPath(config.Root, fmt.Sprintf(setup_xml_path, storageType)).Children
	vars := make(map[string]string)
	for _, node := range nodes {
		if node.Name.Local != "" {
			vars[node.Name.Local] = node.GetValue()
		}
	}
	return vars
}

func getNodesByPath(node *xml.Node, path string) []*xml.Node {
	i := strings.Index(path, "/")
	if i >= 0 {
		name, path := path[:i], path[i+1:]
		node = node.SelectNode("", name)
		return getNodesByPath(node, path)
	} else {
		return node.SelectNodes("", path)
	}
}

func getNodeByPath(node *xml.Node, path string) *xml.Node {
	nodes := getNodesByPath(node, path)
	if len(nodes) > 0 {
		return nodes[0]
	} else {
		return nil
	}
}
