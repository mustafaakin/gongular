package gongular

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"net/http"
)

func TestContext_Fail(t *testing.T) {
	c := Context{}
	c.Fail(http.StatusBadRequest, "hello world")

	assert.Equal(t, c.status, http.StatusBadRequest)
	assert.Equal(t, c.bodyInterface, "hello world")
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

func TestContext_SetBody(t *testing.T) {
	c := Context{}

	b := make([]byte, 30)
	c.SetBody(b)

	// These unit-tests are getting ridiculous but trust me they will get better
	assert.Equal(t, b, c.body)
}