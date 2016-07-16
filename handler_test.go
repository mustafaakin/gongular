package gongular

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewRouter(t *testing.T) {
	r := NewRouter()

	assert.NotNil(t, r.DebugLog)
	assert.NotNil(t, r.InfoLog)

	//	assert.Equal(t, DefaultErrorHandle, r.ErrorHandler)
	assert.NotNil(t, r.router)
	assert.Equal(t, "", r.prefix)
	assert.NotNil(t, r.injector)
}
