package gongular

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"reflect"
	"testing"
)

func TestNewInjector(t *testing.T) {
	i := NewInjector()

	assert.NotNil(t, i.customProviders)
	assert.NotNil(t, i.values)
}

func TestInjector_Provide(t *testing.T) {
	i := NewInjector()

	// Some interface
	type SomeInterface struct {
		Username string
		Age      int
	}

	// pointer to it
	s := &SomeInterface{
		Username: "bla-bla",
		Age:      3,
	}
	i.Provide(s)

	name := reflect.TypeOf(s)

	assert.NotNil(t, i.values[name])
	assert.Equal(t, i.values[name], s)
}

func TestInjector_ProvideCustom(t *testing.T) {
	i := NewInjector()

	type SomeInterface struct {
		Username string
		Age      int
	}

	i.ProvideCustom(&SomeInterface{}, func(w http.ResponseWriter, r *http.Request) (error, interface{}) {
		return nil, &SomeInterface{
			Username: "bla-bla",
			Age:      3,
		}
	})

	// TODO: Not complete yet
}
