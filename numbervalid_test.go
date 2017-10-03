package gongular

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckIntRangeInt8(t *testing.T) {
	var i int64
	i = 30
	bool, _, _ := checkIntRange(reflect.Int8, i)
	assert.True(t, bool)
}

func TestCheckIntNotRangeInt8(t *testing.T) {
	var i int64
	i = 300
	bool, _, _ := checkIntRange(reflect.Int8, i)
	assert.False(t, bool)
}

func TestCheckIntRangeInt16(t *testing.T) {
	var i int64
	i = 4398
	bool, _, _ := checkIntRange(reflect.Int16, i)
	assert.True(t, bool)
}

func TestCheckIntNotRangeInt16(t *testing.T) {
	var i int64
	i = 30000000
	bool, _, _ := checkIntRange(reflect.Int16, i)
	assert.False(t, bool)
}

func TestCheckIntRangeInt(t *testing.T) {
	var i int64
	i = 912389012
	bool, _, _ := checkIntRange(reflect.Int32, i)
	assert.True(t, bool)
}

func TestCheckIntNotRangeInt(t *testing.T) {
	var i int64
	i = 30000000000000
	bool, _, _ := checkIntRange(reflect.Int32, i)
	assert.False(t, bool)
}

func TestCheckIntRangeInt64(t *testing.T) {
	var i int64
	i = 912389012132
	bool, _, _ := checkIntRange(reflect.Int64, i)
	assert.True(t, bool)
}
