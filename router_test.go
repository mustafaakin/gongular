package gongular

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"fmt"
)

func getBasicResponse(t *testing.T, r *Router, path string) (int, string) {
	resp := httptest.NewRecorder()

	uri := path

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		t.Fatal(err)
	}

	r.GetHandler().ServeHTTP(resp, req)
	if p, err := ioutil.ReadAll(resp.Body); err != nil {
		t.Fail()
		return resp.Code, ""
	} else {
		return resp.Code, string(p)
	}
}

func TestRouter_DisableDebug(t *testing.T) {
	r := NewRouter()
	r.DisableDebug()

	assert.Equal(t, 0, r.DebugLog.Flags())
}

func TestRouter_EnableDebug(t *testing.T) {
	r := NewRouter()
	r.EnableDebug()

	assert.Equal(t, log.LstdFlags, r.DebugLog.Flags())
}

func TestRouter_GET_string(t *testing.T) {
	r := NewRouterTest()
	r.GET("/", func() string {
		return "TEST"
	})

	code, content := getBasicResponse(t, r, "/")
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "\"TEST\"", content)
}

func TestRouter_GET_bool(t *testing.T) {
	r := NewRouterTest()
	r.GET("/", func() bool {
		return true
	})

	code, content := getBasicResponse(t, r, "/")
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "true", content)
}

func TestRouter_GET_status(t *testing.T) {
	r := NewRouterTest()
	r.GET("/", func(c *Context) {
		c.Status(http.StatusNetworkAuthenticationRequired)
	})

	code, _ := getBasicResponse(t, r, "/")
	assert.Equal(t, http.StatusNetworkAuthenticationRequired, code)
}

func TestRouter_GET_param_string(t *testing.T) {
	r := NewRouterTest()

	type TestParam struct {
		UserId string
	}

	const UserId = "304050ABCDEF"

	r.GET("/user/:UserId", func(p TestParam) string {
		assert.Equal(t, UserId, p.UserId)
		return p.UserId
	})

	p := "/user/" + UserId

	code, content := getBasicResponse(t, r, p)

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "\"" + UserId + "\"", content)
}

func TestRouter_GET_param_int(t *testing.T) {
	r := NewRouterTest()

	type TestParam struct {
		UserId int
	}

	const UserId = 227

	r.GET("/user/:UserId", func(p TestParam) int {
		assert.Equal(t, UserId, p.UserId)
		return p.UserId
	})

	p := fmt.Sprintf("/user/%d", UserId)

	code, content := getBasicResponse(t, r, p)

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, fmt.Sprintf("%d", UserId), content)
}