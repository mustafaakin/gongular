package gongular

import (
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

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
