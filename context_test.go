package gongular

import (
	"io/ioutil"
	"log"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContext_Fail(t *testing.T) {
	c := Context{}
	c.Fail(http.StatusBadRequest, "hello world")

	assert.Equal(t, c.status, http.StatusBadRequest)
	assert.Equal(t, c.body, "hello world")
}

func TestContext_Header(t *testing.T) {
	c := Context{}
	c.headers = make(map[string]string)

	c.Header("Content-type", "application/json")
	assert.Equal(t, c.headers["Content-type"], "application/json")
}

func TestContext_Request(t *testing.T) {
	c := Context{}
	req := &http.Request{}
	c.r = req

	assert.Equal(t, req, c.Request())
}

func TestContext_Status(t *testing.T) {
	c := Context{}
	c.Status(http.StatusTeapot)
	assert.Equal(t, http.StatusTeapot, c.status)
}

func TestContext_StatusTwice(t *testing.T) {
	c := Context{}
	// TODO: Remove logger
	c.logger = log.New(ioutil.Discard, "", 0)

	c.Status(http.StatusTeapot)
	assert.Equal(t, http.StatusTeapot, c.status)
	c.Status(http.StatusInternalServerError)
	assert.Equal(t, http.StatusTeapot, c.status)
}
