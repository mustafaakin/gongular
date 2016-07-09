package gongular

import (
	"encoding/json"
	"log"
	"net/http"
)

// Context is an object that is alive during an HTTP Request. It holds useful information about a request and allows
// the gongular to hold the information, then serialize it to the client whenever all handlers are finished.
type Context struct {
	r             *http.Request
	w             http.ResponseWriter
	status        int
	headers       map[string]string
	body          []byte
	bodyInterface interface{}
	logger        *log.Logger
	stopChain     bool
}

// ContextFromRequest creates a new Context object from a valid  HTTP Request.
func ContextFromRequest(w http.ResponseWriter, r *http.Request, logger *log.Logger) *Context {
	return &Context{
		r:       r,
		w:       w,
		headers: make(map[string]string),
		logger:  logger,
	}
}

// Request returns the request object so that it can be used in middlewares or handlers.
func (c *Context) Request() *http.Request {
	return c.r
}

// Status sets the resposne code for a request. It generates a warning if it has been tried to set multiple times.
func (c *Context) Status(status int) {
	// Meaning no status written before
	if c.status == 0 {
		c.status = status
	} else {
		c.logger.Printf("Tried to set request status '%d' but it was previously set to '%d'\n", status, c.status)
	}
}

// StopChain marks the context as chain is going to be stopped, meaning no other handlers will be executed.
func (c *Context) StopChain() {
	c.stopChain = true
}

// Header sets an HTTP header for a given key and value
func (c *Context) Header(key, value string) {
	c.headers[key] = value
}

// SetBody sets the body as a byte
func (c *Context) SetBody(b []byte) {
	c.body = b
}

// SetBodyJSON sets the given interface as to be
func (c *Context) SetBodyJSON(v interface{}) {
	c.bodyInterface = v
}

// Fail stops the chain with a status code and an object
func (c *Context) Fail(status int, msg interface{}) {
	c.StopChain()
	c.Status(status)
	c.SetBodyJSON(msg)
}

// finalize writes HTTP status code, headers and the body.
func (c *Context) finalize() int {
	if c.status == 0 {
		c.status = http.StatusOK
	}

	c.w.WriteHeader(c.status)
	for k, v := range c.headers {
		c.w.Header().Add(k, v)
	}

	// Interface body has precedence over byte body
	if c.bodyInterface != nil {
		c.w.Header().Set("Content-type", "application/json")
		b, _ := json.MarshalIndent(c.bodyInterface, "", "  ")
		bytes, err := c.w.Write(b)
		if err != nil {

		}
		return bytes

	} else if c.body != nil {
		bytes, err := c.w.Write(c.body)
		if err != nil {
			// TODO: What to do with it?
		}
		return bytes
	}
	return 0
}
