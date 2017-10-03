package gongular

import (
	"encoding/json"
	"log"
	"net/http"

	"reflect"

	"github.com/julienschmidt/httprouter"
)

// Context is an object that is alive during an HTTP Request. It holds useful information about a request and allows
// the gongular to hold the information, then serialize it to the client whenever all handlers are finished.
type Context struct {
	r         *http.Request
	w         http.ResponseWriter
	status    int
	headers   map[string]string
	body      interface{}
	logger    *log.Logger
	stopChain bool
	params    httprouter.Params
	path      string

	injectCache map[reflect.Type]map[string]interface{}
}

// ContextFromRequest creates a new Context object from a valid  HTTP Request.
func contextFromRequest(path string, w http.ResponseWriter, r *http.Request, params httprouter.Params, logger *log.Logger) *Context {
	return &Context{
		path:        path,
		r:           r,
		w:           w,
		headers:     make(map[string]string),
		logger:      logger,
		params:      params,
		injectCache: make(map[reflect.Type]map[string]interface{}),
	}
}

// Params returns the URL parameters of the request
func (c *Context) Params() httprouter.Params {
	return c.params
}

// Request returns the request object so that it can be used in middlewares or handlers.
func (c *Context) Request() *http.Request {
	return c.r
}

// Status sets the response code for a request. It generates a warning if it has been tried to set multiple times.
func (c *Context) Status(status int) {
	// Meaning no status written before
	if c.status == 0 {
		c.status = status
	} else {
		c.logger.Printf("Tried to set request status '%d' but it was previously set to '%d'\n", status, c.status)
	}
}

// Logger returns the underlying logger for that specific context
func (c *Context) Logger() *log.Logger {
	return c.logger
}

// MustStatus overrides the status
func (c *Context) MustStatus(status int) {
	c.status = status
}

// StopChain marks the context as chain is going to be stopped, meaning no other handlers will be executed.
func (c *Context) StopChain() {
	c.stopChain = true
}

// Header sets an HTTP header for a given key and value
func (c *Context) Header(key, value string) {
	c.headers[key] = value
}

// SetBody sets the given interface which will be written
func (c *Context) SetBody(v interface{}) {
	c.body = v
}

// Fail stops the chain with a status code and an object
func (c *Context) Fail(status int, msg interface{}) {
	c.StopChain()
	c.Status(status)
	c.SetBody(msg)
}

// Finalize writes HTTP status code, headers and the body.
func (c *Context) Finalize() int {
	if c.status == 0 {
		c.status = http.StatusOK
	}

	for k, v := range c.headers {
		c.w.Header().Set(k, v)
	}

	if c.body != nil {
		if v, ok := c.body.([]byte); ok {
			c.w.WriteHeader(c.status)
			bytes, err := c.w.Write(v)
			if err != nil {
				c.logger.Println("Could not write the response", err)
			}
			return bytes
		}

		b, err := json.MarshalIndent(c.body, "", "  ")
		if err != nil {
			c.logger.Println("Could not serialize the response", err)
			return -1
		}

		c.w.Header().Set("Content-type", "application/json")
		c.w.WriteHeader(c.status)

		bytes, err := c.w.Write(b)
		if err != nil {
			c.logger.Println(err)
		}
		return bytes
	}

	c.w.WriteHeader(c.status)
	return 0
}

func (c *Context) getCachedInjection(tip reflect.Type, key string) (interface{}, bool) {
	if m, ok := c.injectCache[tip]; ok {
		val, ok2 := m[key]
		return val, ok2
	}
	return nil, false
}

func (c *Context) putCachedInjection(tip reflect.Type, key string, val interface{}) {
	if _, ok := c.injectCache[tip]; !ok {
		c.injectCache[tip] = make(map[string]interface{})
	}

	c.injectCache[tip][key] = val
}
