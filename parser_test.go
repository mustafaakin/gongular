package gongular

import (
	"crypto/sha256"
	"net/http"
	"testing"

	"fmt"

	"net/url"

	"bytes"
	"mime/multipart"

	"os"

	"io"

	"io/ioutil"
	"net/http/httptest"

	"strings"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/// Param Tests
type singleParam struct {
	Param struct {
		UserID string
	}
}

func (s *singleParam) Handle(c *Context) error {
	c.SetBody(s.Param.UserID)
	return nil
}

func TestSingleParam(t *testing.T) {
	e := newEngineTest()
	e.GetRouter().GET("/user/:UserID", &singleParam{})

	resp, content := get(t, e, "/user/ahmet")
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, content, "\"ahmet\"")
}

type multiParam struct {
	Param struct {
		UserID string
		Page   int
	}
}

func (m *multiParam) Handle(c *Context) error {
	c.SetBody(fmt.Sprintf("%s:%d", m.Param.UserID, m.Param.Page))
	return nil
}

func TestMultiParam(t *testing.T) {
	e := newEngineTest()
	e.GetRouter().GET("/user/:UserID/page/:Page", &multiParam{})

	resp, content := get(t, e, "/user/ahmet/page/5")
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, content, "\"ahmet:5\"")
}

//////// Query Tests
type queryHandler struct {
	Query struct {
		Age      int
		Name     string
		Favorite string
	}
}

func (q *queryHandler) Handle(c *Context) error {
	c.SetBody(fmt.Sprintf("%d:%s:%s", q.Query.Age, q.Query.Name, q.Query.Favorite))
	return nil
}

func TestQueryHandler(t *testing.T) {
	e := newEngineTest()

	e.GetRouter().GET("/hello", &queryHandler{})
	resp, content := get(t, e, "/hello?Name=mustafa&Age=26&Favorite=blue")

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "\"26:mustafa:blue\"", content)
}

//////// Form Tests
type formHandler struct {
	Form struct {
		Age      int
		Name     string
		Favorite string
		Fraction float64
	}
}

func (q *formHandler) Handle(c *Context) error {
	c.SetBody(fmt.Sprintf("%d:%s:%s:%.2f",
		q.Form.Age, q.Form.Name, q.Form.Favorite, q.Form.Fraction))
	return nil
}

func TestFormHandler(t *testing.T) {
	e := newEngineTest()

	e.GetRouter().POST("/submit", &formHandler{})

	data := url.Values{}
	data.Set("Age", "26")
	data.Set("Name", "mustafa")
	data.Set("Favorite", "blue")
	data.Set("Fraction", "0.34234")

	resp, content := postForm(t, e, "/submit", data)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "\"26:mustafa:blue:0.34\"", content)
}

type formUploadTest struct {
	Form struct {
		SomeFile     *UploadedFile
		RegularValue int
	}
}

func (f *formUploadTest) Handle(c *Context) error {
	s := sha256.New()
	io.Copy(s, f.Form.SomeFile.File)
	resp := fmt.Sprintf("%x:%d", s.Sum(nil), f.Form.RegularValue)
	c.SetBody(resp)
	return nil
}

func TestFormWithFileHandler(t *testing.T) {
	e := newEngineTest()

	e.GetRouter().POST("/upload", &formUploadTest{})

	// Construct the body, by using a multipart writer
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// Open README.md file
	f, err := os.Open("README.md")
	require.NoError(t, err)
	defer f.Close()

	// Copy the file to multipart writer
	fw, err := w.CreateFormFile("SomeFile", "README.md")
	require.NoError(t, err)
	_, err = io.Copy(fw, f)
	require.NoError(t, err)

	// Go to beginning of file to get the sha256 sum
	_, err = f.Seek(0, 0)
	s := sha256.New()
	io.Copy(s, f)

	// Write a regular field
	require.NoError(t, err)
	err = w.WriteField("RegularValue", "35")
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

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, fmt.Sprintf(`"%x:35"`, s.Sum(nil)), content)
}

type bodyTest struct {
	Body struct {
		UserID    int
		Name      string
		Positions []struct {
			X int
			Y int
		}
	}
}

func (b *bodyTest) Handle(c *Context) error {
	s := []string{}
	for _, pos := range b.Body.Positions {
		s = append(s, fmt.Sprintf("%d:%d", pos.X, pos.Y))
	}
	c.SetBody(fmt.Sprintf("%s:%d:%s", b.Body.Name, b.Body.UserID, strings.Join(s, ":")))
	return nil
}

func TestBodyHandler(t *testing.T) {
	e := newEngineTest()

	e.GetRouter().POST("/track", &bodyTest{})

	type pos struct {
		X int
		Y int
	}

	type body struct {
		UserID    int
		Name      string
		Positions []pos
	}

	b := body{
		UserID: 26,
		Name:   "mustafa",
		Positions: []pos{
			{X: 4, Y: 5},
			{X: 9, Y: 50},
		},
	}

	resp, content := post(t, e, "/track", b)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, `"mustafa:26:4:5:9:50"`, content)
}
