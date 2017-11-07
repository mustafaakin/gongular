package gongular

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type simpleHandler struct{}

func (s *simpleHandler) Handle(c *Context) error {
	c.SetBody("selam")
	return nil
}

func TestSimpleGetHandler(t *testing.T) {
	e := newEngineTest()
	e.GetRouter().GET("/", &simpleHandler{})

	resp, content := get(t, e, "/")

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "\"selam\"", content)
}

func TestSimplePostHandler(t *testing.T) {
	e := newEngineTest()
	e.GetRouter().POST("/", &simpleHandler{})

	resp, content := post(t, e, "/", nil)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "\"selam\"", content)
}

func TestSimplePutHandler(t *testing.T) {
	e := newEngineTest()
	e.GetRouter().PUT("/", &simpleHandler{})

	resp, content := respWrap(t, e, "/", http.MethodPut, nil)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "\"selam\"", content)
}

func TestSimpleHttpMethodsHandler(t *testing.T) {
	e := newEngineTest()

	fns := map[string]func(string, ...RequestHandler){
		http.MethodHead:    e.GetRouter().HEAD,
		http.MethodConnect: e.GetRouter().CONNECT,
		http.MethodDelete:  e.GetRouter().DELETE,
		http.MethodOptions: e.GetRouter().OPTIONS,
		http.MethodPatch:   e.GetRouter().PATCH,
		http.MethodTrace:   e.GetRouter().TRACE,
		http.MethodPut:     e.GetRouter().PUT,
	}

	for method, fn := range fns {
		path := "/" + method
		fn(path, &simpleHandler{})
		resp, content := respWrap(t, e, path, method, nil)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "\"selam\"", content)
	}
}

func TestSimpleHttpMethodsAlternativeHandler(t *testing.T) {
	fns := []string{
		http.MethodHead,
		http.MethodConnect,
		http.MethodDelete,
		http.MethodOptions,
		http.MethodPatch,
		http.MethodTrace,
		http.MethodPut,
	}

	for _, method := range fns {
		e := newEngineTest()

		path := "/" + method
		e.GetRouter().Method(method, path, &simpleHandler{})
		resp, content := respWrap(t, e, path, method, nil)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Equal(t, "\"selam\"", content)
	}
}

type statusSetHandler struct{}

func (s *statusSetHandler) Handle(c *Context) error {
	c.Status(http.StatusExpectationFailed)
	c.Status(http.StatusFailedDependency)
	return nil
}

func TestSetStatus(t *testing.T) {
	e := newEngineTest()
	e.GetRouter().GET("/", &statusSetHandler{})

	resp, _ := get(t, e, "/")
	assert.Equal(t, http.StatusExpectationFailed, resp.Code)
}

type statusMustSetHandler struct{}

func (s *statusMustSetHandler) Handle(c *Context) error {
	c.Status(http.StatusExpectationFailed)
	c.MustStatus(http.StatusTeapot)
	return nil
}

func TestSetMustStatus(t *testing.T) {
	e := newEngineTest()
	e.GetRouter().GET("/", &statusMustSetHandler{})

	resp, _ := get(t, e, "/")
	assert.Equal(t, http.StatusTeapot, resp.Code)
}

type headerHandler struct{}

func (s *headerHandler) Handle(c *Context) error {
	c.Header("X-API-KEY", "selam")
	c.Header("DNT", "1")
	return nil
}

func TestSetHeader(t *testing.T) {
	e := newEngineTest()
	e.GetRouter().GET("/header", &headerHandler{})

	resp, _ := get(t, e, "/header")
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "selam", resp.Header().Get("X-API-KEY"))
	assert.Equal(t, "1", resp.Header().Get("DNT"))
}

type setByteBody struct{}

func (s *setByteBody) Handle(c *Context) error {
	c.SetBody([]byte{20, 30, 40, 60, 243})
	return nil
}

func TestSetByteBody(t *testing.T) {
	e := newEngineTest()
	e.GetRouter().GET("/bytez", &setByteBody{})

	resp, content := respBytes(t, e, "/bytez", http.MethodGet)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, []byte{20, 30, 40, 60, 243}, content)

}
