package storage

import "testing"

type testParam struct{}

func (t *testParam) SetParam(data string) {}

func (t *testParam) GetParam() string {
	return `{"username":"admin", "password":"123456"}`
}

func TestReq(t *testing.T) {
	p := new(testParam)
	h := new(httpRemote)
	h.Init("", nil)
	h.src = "test.txt"
	h.SetStorageParam(p)
	_, _, err := h.call("POST", "/v1.0/token", false)
	// buf := bytes.NewBuffer(nil)
	// io.Copy(buf, res)
	// fmt.Println(buf.String())
	if err != nil {
		t.Error(err.Error())
	}
}
