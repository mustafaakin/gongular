package gongular

import (
	"net/http"
	"testing"

	"database/sql"

	"net/url"

	"fmt"
	"math"

	"bytes"

	"mime/multipart"

	"io/ioutil"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type expectingIntParam struct {
	Param struct {
		UserID int
	}
}

func (e *expectingIntParam) Handle(c *Context) error {
	c.SetBody("WOW")
	return nil
}

func TestIncompatibleParamTypes(t *testing.T) {
	e := newEngineTest()

	e.GetRouter().GET("/hey/:UserID", &expectingIntParam{})

	resp1, content1 := get(t, e, "/hey/5")

	assert.Equal(t, http.StatusOK, resp1.Code)
	assert.Equal(t, `"WOW"`, content1)

	resp2, content2 := get(t, e, "/hey/notReallyAInteger")

	assert.Equal(t, http.StatusBadRequest, resp2.Code)
	assert.NotEqual(t, `"WOW"`, content2)
}

type expectingAlphaNumParam struct {
	Param struct {
		UserID string `valid:"alphanum"`
	}
}

func (e *expectingAlphaNumParam) Handle(c *Context) error {
	c.SetBody("WOW")
	return nil
}

func TestValidationParam(t *testing.T) {
	e := newEngineTest()

	e.GetRouter().GET("/hey/:UserID", &expectingAlphaNumParam{})

	resp1, content1 := get(t, e, "/hey/abc300")

	assert.Equal(t, http.StatusOK, resp1.Code)
	assert.Equal(t, `"WOW"`, content1)

	resp2, content2 := get(t, e, "/hey/abc$")

	assert.Equal(t, http.StatusBadRequest, resp2.Code)
	assert.NotEqual(t, `"WOW"`, content2)
}

type failingInjection struct {
	DB *sql.DB `inject:"primary"`
}

func (e *failingInjection) Handle(c *Context) error {
	c.SetBody("WOW")
	return nil
}

func TestFailingInjection(t *testing.T) {
	e := newEngineTest()

	e.CustomProvideWithKey("primary", &sql.DB{}, func(c *Context) (interface{}, error) {
		return nil, sql.ErrTxDone
	})

	e.GetRouter().GET("/hey", &failingInjection{})

	resp, content := get(t, e, "/hey")

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.NotEqual(t, `"WOW"`, content)
}

type justFail struct{}

func (e *justFail) Handle(c *Context) error {
	c.SetBody("WOW")
	return http.ErrLineTooLong
}

func TestFailing(t *testing.T) {
	e := newEngineTest()

	e.GetRouter().GET("/uuu", &justFail{})

	resp, content := get(t, e, "/uuu")

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.NotEqual(t, `"WOW"`, content)
}

type boundErrors struct {
	Form struct {
		U8  uint8
		I8  int8
		F32 float32
		B   bool
	}
}

func (b *boundErrors) Handle(c *Context) error {
	c.SetBody("WOW")
	return nil
}

func TestBoundErrors_None(t *testing.T) {
	e := newEngineTest()

	e.GetRouter().POST("/uuu", &boundErrors{})

	data := url.Values{}
	data.Set("U8", "240")
	data.Set("I8", "120")
	data.Set("F32", "0.32")
	data.Set("B", "true")

	resp, content := postForm(t, e, "/uuu", data)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, `"WOW"`, content)
}

func TestBoundErrors_Uint(t *testing.T) {
	e := newEngineTest()

	e.GetRouter().POST("/uuu", &boundErrors{})

	data := url.Values{}
	data.Set("U8", "999")
	data.Set("I8", "120")
	data.Set("F32", "0.32")
	data.Set("B", "true")

	resp, content := postForm(t, e, "/uuu", data)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.NotEqual(t, `"WOW"`, content)
}

func TestBoundErrors_Int(t *testing.T) {
	e := newEngineTest()

	e.GetRouter().POST("/uuu", &boundErrors{})

	data := url.Values{}
	data.Set("U8", "240")
	data.Set("I8", "12033")
	data.Set("F32", "0.32")
	data.Set("B", "true")

	resp, content := postForm(t, e, "/uuu", data)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.NotEqual(t, `"WOW"`, content)
}

func TestBoundErrors_Float(t *testing.T) {
	e := newEngineTest()

	e.GetRouter().POST("/uuu", &boundErrors{})

	data := url.Values{}
	data.Set("U8", "240")
	data.Set("I8", "120")
	data.Set("F32", fmt.Sprintf("%f", math.MaxFloat64))
	data.Set("B", "true")

	resp, content := postForm(t, e, "/uuu", data)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.NotEqual(t, `"WOW"`, content)
}

func TestBoundErrors_Bool(t *testing.T) {
	e := newEngineTest()

	e.GetRouter().POST("/uuu", &boundErrors{})

	data := url.Values{}
	data.Set("U8", "240")
	data.Set("I8", "120")
	data.Set("F32", "0.32")
	data.Set("B", "haydaa")

	resp, content := postForm(t, e, "/uuu", data)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.NotEqual(t, `"WOW"`, content)
}

func TestInvalidErrors_Float(t *testing.T) {
	e := newEngineTest()

	e.GetRouter().POST("/uuu", &boundErrors{})

	data := url.Values{}
	data.Set("U8", "240")
	data.Set("I8", "120")
	data.Set("F32", "lol")
	data.Set("B", "true")

	resp, content := postForm(t, e, "/uuu", data)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.NotEqual(t, `"WOW"`, content)
}

func TestInvalidErrors_Int(t *testing.T) {
	e := newEngineTest()

	e.GetRouter().POST("/uuu", &boundErrors{})

	data := url.Values{}
	data.Set("U8", "240")
	data.Set("I8", "lol")
	data.Set("F32", "0.32")
	data.Set("B", "true")

	resp, content := postForm(t, e, "/uuu", data)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.NotEqual(t, `"WOW"`, content)
}

func TestInvalidErrors_UInt(t *testing.T) {
	e := newEngineTest()

	e.GetRouter().POST("/uuu", &boundErrors{})

	data := url.Values{}
	data.Set("U8", "lol")
	data.Set("I8", "120")
	data.Set("F32", "0.32")
	data.Set("B", "true")

	resp, content := postForm(t, e, "/uuu", data)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.NotEqual(t, `"WOW"`, content)
}

type dummyBodyHandler struct {
	Body struct {
		Name string
		Age  int
	}
}

func (d *dummyBodyHandler) Handle(c *Context) error {
	c.SetBody("wow")
	return nil
}

func TestInvalidErrors_Body(t *testing.T) {
	e := newEngineTest()

	e.GetRouter().POST("/uuu", &dummyBodyHandler{})

	b := bytes.NewBufferString("not a jason")
	resp, content := respWrap(t, e, "/uuu", http.MethodPost, b)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.NotEqual(t, `"WOW"`, content)
}

type dummyFileHandler struct {
	Form struct {
		File1             *UploadedFile
		NotARelevantValue int
	}
}

func (d *dummyFileHandler) Handle(c *Context) error {
	c.SetBody("dummyFileHandler")
	return nil
}

func TestInvalidErrors_NoFileSupplied(t *testing.T) {
	e := newEngineTest()

	e.GetRouter().POST("/upload", &dummyFileHandler{})

	// Construct the body, by using a multipart writer
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	err := w.WriteField("NotARelevantValue", "35")
	assert.NoError(t, err)
	w.Close()

	// Manual construction of the request
	resp := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodPost, "/upload", &b)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	e.GetHandler().ServeHTTP(resp, req)
	p, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	content := string(p)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.NotEqual(t, `"WOW"`, content)
}
