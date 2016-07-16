package gongular

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func resp_wrap(t *testing.T, r *Router, path, method string, reader io.Reader) (*httptest.ResponseRecorder, string) {
	resp := httptest.NewRecorder()

	uri := path

	req, err := http.NewRequest(method, uri, reader)
	if err != nil {
		t.Fatal(err)
	}

	r.GetHandler().ServeHTTP(resp, req)
	if p, err := ioutil.ReadAll(resp.Body); err != nil {
		t.Fail()
		return resp, ""
	} else {
		return resp, string(p)
	}
}

func get(t *testing.T, r *Router, path string) (*httptest.ResponseRecorder, string) {
	return resp_wrap(t, r, path, "GET", nil)
}

func post(t *testing.T, r *Router, path string, body interface{}) (*httptest.ResponseRecorder, string) {
	if body != nil {
		b, err := json.Marshal(body)
		assert.NoError(t, err)
		return resp_wrap(t, r, path, "POST", bytes.NewBuffer(b))
	} else {
		return resp_wrap(t, r, path, "POST", nil)
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

	resp, content := get(t, r, "/")
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "\"TEST\"", content)
}

func TestRouter_GET_bool(t *testing.T) {
	r := NewRouterTest()
	r.GET("/", func() bool {
		return true
	})

	resp, content := get(t, r, "/")
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "true", content)
}

func TestRouter_GET_status(t *testing.T) {
	r := NewRouterTest()
	r.GET("/", func(c *Context) {
		c.Status(http.StatusNetworkAuthenticationRequired)
	})

	resp, _ := get(t, r, "/")
	assert.Equal(t, http.StatusNetworkAuthenticationRequired, resp.Code)
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

	resp, content := get(t, r, p)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "\""+UserId+"\"", content)
}

func TestRouter_GET_param_string_validation(t *testing.T) {
	r := NewRouterTest()

	type TestParam struct {
		UserId string `valid:"alphanum"`
	}

	const UserId = "!!!AAAA"

	r.GET("/user/:UserId", func(p TestParam) {
		assert.Equal(t, UserId, p.UserId)
	})

	p := "/user/" + UserId

	resp, _ := get(t, r, p)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
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

	resp, content := get(t, r, p)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, fmt.Sprintf("%d", UserId), content)
}

func TestRouter_GET_query(t *testing.T) {
	r := NewRouterTest()

	type TestQuery struct {
		UserId int
		Name   string
	}

	const UserId = 227
	const Name = "mustafa-mistik"

	r.GET("/hello", func(q TestQuery) {
		assert.Equal(t, UserId, q.UserId)
		assert.Equal(t, Name, q.Name)
	})

	u, err := url.Parse("/hello")
	assert.Nil(t, err)
	q := u.Query()
	q.Set("UserId", fmt.Sprintf("%d", UserId))
	q.Set("Name", Name)

	u.RawQuery = q.Encode()

	resp, _ := get(t, r, u.String())

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestRouter_GET_query_validate(t *testing.T) {
	r := NewRouterTest()

	type TestQuery struct {
		UserId int
		Name   string `valid:"alphanum"`
	}

	const UserId = 227
	const Name = "mustafa-mistik"

	r.GET("/hello", func(q TestQuery) {
		assert.Equal(t, UserId, q.UserId)
		assert.Equal(t, Name, q.Name)
	})

	u, err := url.Parse("/hello")
	assert.Nil(t, err)
	q := u.Query()
	q.Set("UserId", fmt.Sprintf("%d", UserId))
	q.Set("Name", Name)

	u.RawQuery = q.Encode()

	resp, _ := get(t, r, u.String())

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestRouter_POST_basic(t *testing.T) {
	r := NewRouterTest()

	const RESPONSE = "hello world 123"

	r.POST("/hello", func() string {
		return RESPONSE
	})

	resp, content := post(t, r, "/hello", nil)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, fmt.Sprintf(`"%s"`, RESPONSE), content)
}

func TestRouter_POST_try_get(t *testing.T) {
	r := NewRouterTest()

	const RESPONSE = "hello world 123"

	r.POST("/hello", func() string {
		return RESPONSE
	})

	resp, _ := get(t, r, "/hello")

	assert.Equal(t, http.StatusMethodNotAllowed, resp.Code)
}

func TestRouter_POST_body(t *testing.T) {
	r := NewRouterTest()

	type TestBody struct {
		Username string
		Age      int
	}

	BODY := TestBody{
		Username: "mustafa",
		Age:      25,
	}

	r.POST("/hello", func(b TestBody) {
		assert.Equal(t, BODY.Username, b.Username)
		assert.Equal(t, BODY.Age, b.Age)
	})

	resp, _ := post(t, r, "/hello", BODY)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestRouter_POST_body_fail(t *testing.T) {
	r := NewRouterTest()

	type TestBody struct {
		Username string
		Age      int
	}

	BODY := TestBody{
		Username: "mustafa",
		Age:      25,
	}

	r.POST("/hello", func(b TestBody) {
		// Should not be here
		assert.NotEqual(t, 1, 1)
		assert.NotEqual(t, BODY.Username, b.Username)
		assert.NotEqual(t, BODY.Age, b.Age)
	})

	resp, _ := post(t, r, "/hello", "IN-VALIDJSON}")

	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestRouter_Group(t *testing.T) {
	r := NewRouterTest()

	r.GET("/", func() string {
		return "index"
	})

	g := r.Group("/admin", func(c *Context) {
		assert.True(t, strings.HasPrefix(c.Request().URL.String(), "/admin/"))
	})

	g.GET("/get-page", func() string {
		return "get-admin-page"
	})

	g.POST("/post-page", func() int {
		return 5
	})

	sg := g.Group("/sub", func(c *Context) {
		assert.True(t, strings.HasPrefix(c.Request().URL.String(), "/admin/sub/"))
	})

	sg.GET("/wow", func() string {
		return "much request"
	})

	// Make requests and test

	resp1, content1 := get(t, r, "/")
	assert.Equal(t, http.StatusOK, resp1.Code)
	assert.Equal(t, `"index"`, content1)

	resp2, content2 := get(t, r, "/admin/get-page")
	assert.Equal(t, http.StatusOK, resp2.Code)
	assert.Equal(t, `"get-admin-page"`, content2)

	resp3, _ := get(t, r, "/admin")
	assert.Equal(t, http.StatusNotFound, resp3.Code)

	resp4, content4 := post(t, r, "/admin/post-page", nil)
	assert.Equal(t, http.StatusOK, resp4.Code)
	assert.Equal(t, `5`, content4)

	resp5, content5 := get(t, r, "/admin/sub/wow")
	assert.Equal(t, http.StatusOK, resp5.Code)
	assert.Equal(t, `"much request"`, content5)
}

func TestRouter_Error(t *testing.T) {
	r := NewRouterTest()

	err := errors.New("error occurred sorry")

	r.GET("/fail", func() (string, error) {
		return "wow-much-request", err
	})

	resp, content := get(t, r, "/fail")

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.NotEqual(t, `"wow-much-request"`, content)
}

func TestRouter_Provide(t *testing.T) {
	r := NewRouterTest()

	type DB struct {
		Hostname string
		Password string
	}

	d := &DB{
		Hostname: "mysql-domain.com",
		Password: "1234",
	}

	r.Provide(d)

	r.GET("/provide-test", func(d2 *DB) {
		assert.Equal(t, "mysql-domain.com", d2.Hostname)
		assert.Equal(t, "1234", d2.Password)
		assert.Equal(t, d, d2)
	})

	resp, _ := get(t, r, "/provide-test")
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestRouter_CustomProvide(t *testing.T) {
	r := NewRouterTest()

	type DB struct {
		Hostname string
		Password string
	}

	r.ProvideCustom(&DB{}, func(c *Context) (error, interface{}) {
		return nil, &DB{
			Hostname: "mysql-domain.com",
			Password: "1234",
		}
	})

	r.GET("/custom-provide-test", func(d2 *DB) {
		assert.Equal(t, "mysql-domain.com", d2.Hostname)
		assert.Equal(t, "1234", d2.Password)
	})

	resp, _ := get(t, r, "/custom-provide-test")
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestRouter_CustomProvideError(t *testing.T) {
	r := NewRouterTest()

	type DB struct {
		Hostname string
		Password string
	}

	r.ProvideCustom(&DB{}, func(c *Context) (error, interface{}) {
		return errors.New("Cannot provide sorry"), nil
	})

	r.GET("/custom-provide-err", func(d *DB) {
		// Wow, even if we are here the d should be null
		assert.Nil(t, d)

		// We should not even be here
		assert.NotEqual(t, 1, 1)
	})

	resp, _ := get(t, r, "/custom-provide-err")
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
}

func TestRouter_Provide_Unknown(t *testing.T) {
	r := NewRouterTest()

	type DB struct {
		Hostname string
		Password string
	}

	assert.Panics(t, func() {
		r.GET("/custom-provide-err", func(d *DB) {
			// Wow, even if we are here the d should be null
			assert.Nil(t, d)

			// We should not even be here
			assert.NotEqual(t, 1, 1)
		})

		resp, _ := get(t, r, "/custom-provide-unknown")
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
	})
}

func TestRouter_CustomProvide_Nil(t *testing.T) {
	r := NewRouterTest()

	type DB struct {
		Hostname string
		Password string
	}

	r.ProvideCustom(&DB{}, func(c *Context) (error, interface{}) {
		// NO error returned
		return nil, nil
	})

	r.GET("/custom-provide-nil", func(d *DB) {
		assert.Nil(t, d)
	})

	resp, _ := get(t, r, "/custom-provide-nil")
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestRouter_GET_header(t *testing.T) {
	r := NewRouterTest()

	r.GET("/header", func(c *Context) {
		c.Header("abc", "123")
		c.Header("def", "456")
	})

	resp, _ := get(t, r, "/header")
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, resp.Header().Get("abc"), "123")
	assert.Equal(t, resp.Header().Get("def"), "456")
}

func TestRouter_NoPanic(t *testing.T) {
	r := NewRouterTest()

	assert.NotPanics(t, func(){
		r.GET("/panic", func() string{
			panic("haydaa")
			return "sorry"
		})

		resp, _ := get(t, r, "/panic")
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
	})
}