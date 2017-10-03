package gongular

import (
	"database/sql"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type injectionDirectHandler struct {
	Param struct {
		UserID uint
	}
	Database *sql.DB
}

func (i *injectionDirectHandler) Handle(c *Context) error {
	c.SetBody(fmt.Sprintf("%p:%d", i.Database, i.Param.UserID))
	return nil
}

func TestInjectDirect(t *testing.T) {
	e := newEngineTest()
	db := new(sql.DB)
	e.Provide(db)

	e.GetRouter().GET("/my/db/interaction/:UserID", &injectionDirectHandler{})

	resp, content := get(t, e, "/my/db/interaction/5")

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, fmt.Sprintf(`"%p:5"`, db), content)
}

type injectKey struct {
	Val1 int `inject:"val1"`
	Val2 int `inject:"val2"`
}

func (i *injectKey) Handle(c *Context) error {
	c.SetBody(i.Val1 * i.Val2)
	return nil
}

func TestInjectKey(t *testing.T) {
	e := newEngineTest()
	e.ProvideWithKey("val1", 71)
	e.ProvideWithKey("val2", 97)

	e.GetRouter().GET("/", &injectKey{})

	resp, content := get(t, e, "/")

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, `6887`, content)
}

type injectCustom struct {
	DB *sql.DB
}

func (i *injectCustom) Handle(c *Context) error {
	c.SetBody(fmt.Sprintf("%p", i.DB))
	return nil
}

func TestInjectCustom(t *testing.T) {
	e := newEngineTest()

	var d *sql.DB
	e.CustomProvide(&sql.DB{}, func(c *Context) (interface{}, error) {
		d = new(sql.DB)
		return d, nil
	})

	e.GetRouter().GET("/", &injectCustom{})

	resp1, content1 := get(t, e, "/")
	assert.Equal(t, http.StatusOK, resp1.Code)
	assert.Equal(t, fmt.Sprintf(`"%p"`, d), content1)

	// Again
	resp2, content2 := get(t, e, "/")
	assert.Equal(t, http.StatusOK, resp2.Code)
	assert.Equal(t, fmt.Sprintf(`"%p"`, d), content2)
}

type injectCustomCache1 struct {
	DB *sql.DB
}
type injectCustomCache2 struct {
	DB *sql.DB
}

func (i *injectCustomCache1) Handle(c *Context) error {
	c.logger.Printf("%p", i.DB)
	return nil
}

func (i *injectCustomCache2) Handle(c *Context) error {
	c.SetBody(fmt.Sprintf("%p", i.DB))
	return nil
}

func TestInjectCustomCache(t *testing.T) {
	e := newEngineTest()

	var d *sql.DB
	callCount := 0
	e.CustomProvide(&sql.DB{}, func(c *Context) (interface{}, error) {
		d = new(sql.DB)
		callCount++
		return d, nil
	})

	e.GetRouter().GET("/", &injectCustomCache1{}, &injectCustomCache2{})

	resp1, content1 := get(t, e, "/")
	assert.Equal(t, http.StatusOK, resp1.Code)
	assert.Equal(t, fmt.Sprintf(`"%p"`, d), content1)
	assert.Equal(t, 1, callCount)

	// Again
	resp2, content2 := get(t, e, "/")
	assert.Equal(t, http.StatusOK, resp2.Code)
	assert.Equal(t, fmt.Sprintf(`"%p"`, d), content2)

	assert.Equal(t, 2, callCount)
}
