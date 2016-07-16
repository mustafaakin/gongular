package gongular

import (
	"reflect"
)

// Injector remembers the provided values so that you can inject whenever
// you need them
type Injector struct {
	values          map[reflect.Type]interface{}
	customProviders map[reflect.Type]CustomProvideFunction
}

// NewInjector creates an Injector with its initial structures initialized
func NewInjector() *Injector {
	return &Injector{
		values:          make(map[reflect.Type]interface{}),
		customProviders: make(map[reflect.Type]CustomProvideFunction),
	}
}

// Provide registers given value depending on its name
func (inj *Injector) Provide(value interface{}) {
	name := reflect.TypeOf(value)
	inj.values[name] = value
}

// ProvideCustom gets the type information from value, however calls CustomProvideFunction
// each time to provide when needed
func (inj *Injector) ProvideCustom(value interface{}, fn CustomProvideFunction) {
	name := reflect.TypeOf(value)
	inj.customProviders[name] = fn
}

// CustomProvideFunction is called whenever a value is needed to be provided
// with custom logic
type CustomProvideFunction func(c *Context) (error, interface{})
