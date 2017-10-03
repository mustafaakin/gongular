package gongular

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"net/url"

	"github.com/stretchr/testify/assert"
)

type allFields struct {
	U1   uint8
	U2   uint16
	U3   uint32
	U4   uint64
	I1   int8
	I2   int16
	I3   int32
	I4   int64
	F1   float32
	F2   float64
	B1   bool
	B2   bool
	Str1 string
	Str2 string
	Str3 string
}

var set = allFields{
	U1:   4,
	U2:   6312,
	U3:   34314,
	U4:   1023123132,
	I1:   123,
	I2:   1325,
	I3:   12359,
	I4:   495435034,
	F1:   34013.43103024,
	F2:   -1320213995243.44353243103024,
	B1:   true,
	B2:   false,
	Str1: "selambro",
	Str2: "wassup",
	Str3: "bye",
}

type complexHandler1 struct {
	Param allFields
	Body  allFields
	Query allFields
}

func (s *complexHandler1) Handle(c *Context) error {
	b1 := reflect.DeepEqual(s.Param, set)
	b2 := reflect.DeepEqual(s.Body, set)
	b3 := reflect.DeepEqual(s.Query, set)

	c.SetBody(fmt.Sprintf("%t:%t:%t", b1, b2, b3))
	return nil
}

func TestComplex1Handler(t *testing.T) {
	e := newEngineTest()
	e.GetRouter().POST("/complex/:U1/:U2/:U3/:U4/:I1/:I2/:I3/:I4/:F1/:F2/:B1/:B2/:Str1/:Str2/:Str3",
		&complexHandler1{})

	// Construct the path params
	path := fmt.Sprintf(
		"/complex/%d/%d/%d/%d/%d/%d/%d/%d/%f/%f/%t/%t/%s/%s/%s",
		set.U1, set.U2, set.U3, set.U4, set.I1, set.I2, set.I3, set.I4, set.F1, set.F2,
		set.B1, set.B2, set.Str1, set.Str2, set.Str3,
	)

	u, err := url.Parse(path)
	assert.Nil(t, err)

	// Construct the query params
	q := u.Query()
	tip := reflect.TypeOf(set)
	val := reflect.ValueOf(set)
	for i := 0; i < tip.NumField(); i++ {
		q.Set(tip.Field(i).Name, fmt.Sprint(val.Field(i).Interface()))
	}

	u.RawQuery = q.Encode()

	// Add body and post
	resp, content := post(t, e, u.String(), set)

	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, `"true:true:true"`, content)
}
