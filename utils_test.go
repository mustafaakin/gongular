package gongular

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"net/url"

	"github.com/stretchr/testify/assert"
)

func respBytes(t *testing.T, e *Engine, path, method string) (*httptest.ResponseRecorder, []byte) {
	resp := httptest.NewRecorder()

	uri := path

	req, err := http.NewRequest(method, uri, nil)
	if err != nil {
		t.Fatal(err)
	}

	e.GetHandler().ServeHTTP(resp, req)
	p, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fail()
		return resp, nil
	}
	return resp, p
}

func respWrap(t *testing.T, e *Engine, path, method string, reader io.Reader) (*httptest.ResponseRecorder, string) {
	resp := httptest.NewRecorder()

	uri := path

	req, err := http.NewRequest(method, uri, reader)
	if err != nil {
		t.Fatal(err)
	}

	e.GetHandler().ServeHTTP(resp, req)
	p, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fail()
		return resp, ""
	}
	return resp, string(p)
}

func get(t *testing.T, e *Engine, path string) (*httptest.ResponseRecorder, string) {
	return respWrap(t, e, path, "GET", nil)
}

func post(t *testing.T, e *Engine, path string, body interface{}) (*httptest.ResponseRecorder, string) {
	if body != nil {
		b, err := json.Marshal(body)
		assert.NoError(t, err)
		return respWrap(t, e, path, "POST", bytes.NewBuffer(b))
	}
	return respWrap(t, e, path, "POST", nil)
}

func postForm(t *testing.T, e *Engine, path string, values url.Values) (*httptest.ResponseRecorder, string) {
	resp := httptest.NewRecorder()

	uri := path

	req, err := http.NewRequest(http.MethodPost, uri, bytes.NewBufferString(values.Encode()))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	e.GetHandler().ServeHTTP(resp, req)
	p, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fail()
		return resp, ""
	}
	return resp, string(p)
}

func newEngineTest() *Engine {
	e := NewEngine()
	e.SetRouteCallback(NoOpRouteCallback)
	return e
}
