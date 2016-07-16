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

func resp_wrap(t *testing.T, r *Router, path, method string, reader io.Reader) (int, string) {
	resp := httptest.NewRecorder()

	uri := path

	req, err := http.NewRequest(method, uri, reader)
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

func get(t *testing.T, r *Router, path string) (int, string) {
	return resp_wrap(t, r, path, "GET", nil)
}

func post(t *testing.T, r *Router, path string, body interface{}) (int, string) {
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

	code, content := get(t, r, "/")
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "\"TEST\"", content)
}

func TestRouter_GET_bool(t *testing.T) {
	r := NewRouterTest()
	r.GET("/", func() bool {
		return true
	})

	code, content := get(t, r, "/")
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, "true", content)
}

func TestRouter_GET_status(t *testing.T) {
	r := NewRouterTest()
	r.GET("/", func(c *Context) {
		c.Status(http.StatusNetworkAuthenticationRequired)
	})

	code, _ := get(t, r, "/")
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

	code, content := get(t, r, p)

	assert.Equal(t, http.StatusOK, code)
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

	code, _ := get(t, r, p)
	assert.Equal(t, http.StatusBadRequest, code)
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

	code, content := get(t, r, p)

	assert.Equal(t, http.StatusOK, code)
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

	code, _ := get(t, r, u.String())

	assert.Equal(t, http.StatusOK, code)
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

	code, _ := get(t, r, u.String())

	assert.Equal(t, http.StatusBadRequest, code)
}

func TestRouter_POST_basic(t *testing.T) {
	r := NewRouterTest()

	const RESPONSE = "hello world 123"

	r.POST("/hello", func() string {
		return RESPONSE
	})

	code, content := post(t, r, "/hello", nil)

	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, fmt.Sprintf(`"%s"`, RESPONSE), content)
}

func TestRouter_POST_try_get(t *testing.T) {
	r := NewRouterTest()

	const RESPONSE = "hello world 123"

	r.POST("/hello", func() string {
		return RESPONSE
	})

	code, _ := get(t, r, "/hello")

	assert.Equal(t, http.StatusMethodNotAllowed, code)
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

	code, _ := post(t, r, "/hello", BODY)

	assert.Equal(t, http.StatusOK, code)
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

	code, _ := post(t, r, "/hello", "IN-VALIDJSON}")

	assert.Equal(t, http.StatusBadRequest, code)
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

	code1, content1 := get(t, r, "/")
	assert.Equal(t, http.StatusOK, code1)
	assert.Equal(t, `"index"`, content1)

	code2, content2 := get(t, r, "/admin/get-page")
	assert.Equal(t, http.StatusOK, code2)
	assert.Equal(t, `"get-admin-page"`, content2)

	code3, _ := get(t, r, "/admin")
	assert.Equal(t, http.StatusNotFound, code3)

	code4, content4 := post(t, r, "/admin/post-page", nil)
	assert.Equal(t, http.StatusOK, code4)
	assert.Equal(t, `5`, content4)

	code5, content5 := get(t, r, "/admin/sub/wow")
	assert.Equal(t, http.StatusOK, code5)
	assert.Equal(t, `"much request"`, content5)
}

func TestRouter_Error(t *testing.T) {
	r := NewRouterTest()

	err := errors.New("error occurred sorry")

	r.GET("/fail", func() (string, error) {
		return "wow-much-request", err
	})

	code, content := get(t, r, "/fail")

	assert.Equal(t, http.StatusInternalServerError, code)
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

	r.GET("/provide-test", func(d2 *DB){
		assert.Equal(t, "mysql-domain.com", d2.Hostname)
		assert.Equal(t, "1234", d2.Password)
		assert.Equal(t, d, d2)
	})

	code, _ := get(t,r, "/provide-test")
	assert.Equal(t, http.StatusOK, code)
}

func TestRouter_CustomProvide(t *testing.T) {
	r := NewRouterTest()

	type DB struct {
		Hostname string
		Password string
	}

	r.ProvideCustom(&DB{}, func(c *Context) (error, interface{}){
		return nil, &DB{
			Hostname: "mysql-domain.com",
			Password: "1234",
		}
	})

	r.GET("/custom-provide-test", func(d2 *DB){
		assert.Equal(t, "mysql-domain.com", d2.Hostname)
		assert.Equal(t, "1234", d2.Password)
	})

	code, _ := get(t,r, "/custom-provide-test")
	assert.Equal(t, http.StatusOK, code)
}

func TestRouter_CustomProvideError(t *testing.T) {
	r := NewRouterTest()

	type DB struct {
		Hostname string
		Password string
	}

	r.ProvideCustom(&DB{}, func(c *Context) (error, interface{}){
		return errors.New("Cannot provide sorry"), nil
	})

	r.GET("/custom-provide-err", func(d *DB){
		// Wow, even if we are here the d should be null
		assert.Nil(t, d)

		// We should not even be here
		assert.NotEqual(t, 1, 1)
	})

	code, _ := get(t,r, "/custom-provide-err")
	assert.Equal(t, http.StatusInternalServerError, code)
}

func TestRouter_Provide_Unknown(t *testing.T) {
	r := NewRouterTest()

	type DB struct {
		Hostname string
		Password string
	}

	assert.Panics(t, func(){
		r.GET("/custom-provide-err", func(d *DB){
			// Wow, even if we are here the d should be null
			assert.Nil(t, d)

			// We should not even be here
			assert.NotEqual(t, 1, 1)
		})


		code, _ := get(t,r, "/custom-provide-unknown")
		assert.Equal(t, http.StatusInternalServerError, code)
	})
}

