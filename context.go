package gongular

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type Context struct {
	r             *http.Request
	w             http.ResponseWriter
	status        int
	headers       map[string]string
	body          []byte
	bodyInterface interface{}
	startedOn     time.Time
	logger        *log.Logger
	stopChain     bool
}

func ContextFromRequest(w http.ResponseWriter, r *http.Request, logger *log.Logger) *Context {
	return &Context{
		r:       r,
		w:       w,
		headers: make(map[string]string),
		logger:  logger,
	}
}

func (c *Context) Request() *http.Request {
	return c.r
}

func (c *Context) Status(status int) {
	// Meaning no status written before
	if c.status == 0 {
		c.status = status
	} else {
		c.logger.Printf("Tried to set request status '%d' but it was previously set to '%d'\n", status, c.status)
	}
}

func (c *Context) StopChain() {
	c.stopChain = true
}

func (c *Context) Header(key, value string) {
	c.headers[key] = value
}

func (c *Context) SetBody(b []byte) {
	c.body = b
}

func (c *Context) SetBodyJSON(v interface{}) {
	c.bodyInterface = v
}

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

func (c *Context) Fail(status int, msg interface{}) {
	c.StopChain()
	c.Status(status)
	c.SetBodyJSON(msg)
}
