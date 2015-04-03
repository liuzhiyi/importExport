package storage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/ddliu/go-httpclient"
)

func init() {
	Register("http", &httpRemote{})
}

type httpRemote struct {
	src       string
	host      string
	writePath string
	readPath  string
	method    string
	scheme    string
	regexp    map[string]string
	config    map[string]string
	colNames  []string
	cursor    int64
	offset    int64
	count     int64
	p         StorageParam
	transport *http.Transport
}

func (h *httpRemote) New() Storage {
	return &httpRemote{}
}

func (h *httpRemote) Init(config map[string]string) {
	h.transport = &http.Transport{}
	h.scheme = config["scheme"]
	h.host = config["host"]
	h.writePath = config["writeUrl"]
	h.readPath = config["readUrl"]
	h.config = make(map[string]string)
	h.config["fileField"] = config["fileField"]
}

func (h *httpRemote) SetConfig(key, val string) error {
	h.config[key] = val
	return nil
}

func (h *httpRemote) SetRegexp(reg map[string]string) {
	h.regexp = reg
}

func (h *httpRemote) SetColNames(cols []string) {
	h.colNames = cols
}

func (h *httpRemote) Read() [][]string {
	var records [][]string

	return records
}

func (h *httpRemote) WriteAll(size int64) {

}

func (h *httpRemote) WriteRow() {

}

func (h *httpRemote) getFileField() string {
	if fileField, ok := h.config["fileField"]; ok {
		return fileField
	} else {
		return "file"
	}
}

func (h *httpRemote) getFileName() string {
	if fileName, ok := h.config["fileName"]; ok {
		return fileName
	} else {
		return "import.csv"
	}
}

func (h *httpRemote) RawWrite(reader io.Reader) {
	//h.call("POST", h.writePath, reader)
	resp, err := httpclient.Post(h.host+h.writePath, map[string]string{
		"@" + h.getFileField(): h.getFileName(),
	})
	content, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(content))
	fmt.Println(err)
}

func (h *httpRemote) Filter(where string) {

}

func (h *httpRemote) SetSize(size int64) {

}

func (h *httpRemote) SetCursor(cursor int64) {
	h.cursor = cursor
}

func (h *httpRemote) postFile(reader io.Reader) (*bytes.Buffer, string) {
	body := bytes.NewBuffer(nil)
	writer := multipart.NewWriter(body)
	defer writer.Close()
	part, _ := writer.CreateFormFile("gdgg", "gxs.csv")
	io.Copy(part, reader)
	ctype := writer.FormDataContentType()
	return body, ctype
}

func (h *httpRemote) getRequest(method, path string, reader io.Reader) (*http.Request, error) {
	var body *bytes.Buffer
	var ctype string

	if reader != nil {
		body, ctype = h.postFile(reader)
	} else {
		body, ctype = h.getBody()
	}
	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "importExport-Client/")
	req.URL.Host = h.host
	req.URL.Scheme = h.scheme

	//ctype := fmt.Sprintf(`multipart/form-data; boundary="%s"`, boundary)
	req.Header.Set("Content-type", ctype)

	return req, nil
}

func (h *httpRemote) SetStorageParam(p StorageParam) {
	h.p = p
}

func (h *httpRemote) getBody() (*bytes.Buffer, string) {
	params := h.p.GetParam()
	buf := bytes.NewBuffer(nil)
	buf.WriteString(params)
	return buf, "text/plain"
}

func (h *httpRemote) call(method, path string, reader io.Reader) (io.ReadCloser, int, error) {
	req, err := h.getRequest(method, path, reader)
	if err != nil {
		return nil, -1, err
	}
	rsp, err := h.HTTPClient().Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			return nil, -1, errors.New("connection refused")
		}

		return nil, -1, fmt.Errorf("An error occurred trying to connect: %v", err)
	}

	if rsp.StatusCode < 200 || rsp.StatusCode >= 400 {
		body, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return nil, -1, err
		}
		if len(body) == 0 {
			return nil, rsp.StatusCode, fmt.Errorf("Error: request returned %s for API route and version %s, check if the server supports the requested API version", http.StatusText(rsp.StatusCode), req.URL)
		}
		return nil, rsp.StatusCode, fmt.Errorf("Error response from daemon: %s", bytes.TrimSpace(body))
	}

	return rsp.Body, rsp.StatusCode, nil
}

func (h *httpRemote) encodeData(data interface{}) (string, error) {
	if data != nil {
		buf, err := json.Marshal(data)
		if err != nil {
			return "", err
		}
		return string(buf), nil
	}
	return "", nil
}

func (h *httpRemote) HTTPClient() *http.Client {
	return &http.Client{Transport: h.transport}
}
