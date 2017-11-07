package gongular

import (
	"errors"
	"net/http"
	"testing"

	"bytes"
	"io/ioutil"

	"github.com/stretchr/testify/assert"
)

type errorTester struct{}

func (e *errorTester) Handle(c *Context) error {
	return errors.New("Shit")
}

func TestEngine_SetRouteCallback(t *testing.T) {
	var isErr error
	fn := func(err error, c *Context) {
		isErr = http.ErrUseLastResponse
	}

	e := newEngineTest()
	e.SetErrorHandler(fn)
	e.GetRouter().GET("/", &errorTester{})

	_, _ = get(t, e, "/")

	assert.Error(t, isErr)
	assert.Equal(t, http.ErrUseLastResponse, isErr)
}

type middlewareFailIfUserId5 struct {
	Param struct {
		UserID int
	}
}

func (m *middlewareFailIfUserId5) Handle(c *Context) error {
	if m.Param.UserID == 5 {
		c.Status(http.StatusTeapot)
		c.SetBody("Sorry")
		c.StopChain()
	}
	return nil
}

func TestGroup(t *testing.T) {
	e := newEngineTest()
	r := e.GetRouter()

	g := r.Group("/api/user/:UserID", &middlewareFailIfUserId5{})
	g.GET("/name", &simpleHandler{})
	g.GET("/wow", &simpleHandler{})

	resp1, _ := get(t, e, "/api/user/30/name")
	assert.Equal(t, http.StatusOK, resp1.Code)

	resp2, _ := get(t, e, "/api/user/5/name")
	assert.Equal(t, http.StatusTeapot, resp2.Code)

	resp3, _ := get(t, e, "/api/user/30/wow")
	assert.Equal(t, http.StatusOK, resp3.Code)

	resp4, _ := get(t, e, "/api/user/5/wow")
	assert.Equal(t, http.StatusTeapot, resp4.Code)
}

func TestEngineWithDefaultRouteCallback(t *testing.T) {
	e := NewEngine()
	e.GetRouter().GET("/", &simpleHandler{})

	resp, content := get(t, e, "/")
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, `"selam"`, content)
}

func TestEngineFileServe(t *testing.T) {
	e := NewEngine()
	e.ServeFile("/", "README.md")
	e.ServeFiles("/static", http.Dir("."))

	bytesReadme, err := ioutil.ReadFile("README.md")
	if err != nil {
		assert.NoError(t, err, "cannot read file")
	}

	// Test binary files as well
	bytesLogo, err := ioutil.ReadFile("logo.png")
	if err != nil {
		assert.NoError(t, err, "cannot read file")
	}

	resp, content := get(t, e, "/")
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, string(bytesReadme), content)

	resp2, content2 := get(t, e, "/static/logo.png")
	assert.Equal(t, http.StatusOK, resp2.Code)

	// Sorry for the inefficiency
	assert.Equal(t, 0, bytes.Compare([]byte(content2), bytesLogo))

	// Not found test
	resp3, _ := get(t, e, "/static/no-file-should-be-here")
	assert.Equal(t, http.StatusNotFound, resp3.Code)

}
