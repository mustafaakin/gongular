package gongular

import (
	"fmt"
	"reflect"
)

// injector remembers the provided values so that you can inject whenever
// you need them
type injector struct {
	unsafeValues    map[string]reflect.Value
	values          map[reflect.Type]map[string]interface{}
	customProviders map[reflect.Type]map[string]CustomProvideFunction
}

// newInjector creates an Injector with its initial structures initialized
func newInjector() *injector {
	return &injector{
		unsafeValues:    make(map[string]reflect.Value),
		values:          make(map[reflect.Type]map[string]interface{}),
		customProviders: make(map[reflect.Type]map[string]CustomProvideFunction),
	}
}

// Provide registers given value depending on its name
func (inj *injector) Provide(value interface{}, key string) {
	tip := reflect.TypeOf(value)
	if inj.values[tip] == nil {
		inj.values[tip] = make(map[string]interface{})
	}
	inj.values[tip][key] = value
}

// ProvideUnsafe registers given value depending on its name
func (inj *injector) ProvideUnsafe(key string, value interface{}) {
	if _, ok := inj.unsafeValues[key]; ok {
		panic(fmt.Sprintf("already provided %#v unsafe value for key %q. ", value, key))
	}

	inj.unsafeValues[key] = reflect.ValueOf(value)
}

// ProvideCustom gets the type information from value, however calls CustomProvideFunction
// each time to provide when needed
func (inj *injector) ProvideCustom(value interface{}, fn CustomProvideFunction, key string) {
	tip := reflect.TypeOf(value)
	if inj.customProviders[tip] == nil {
		inj.customProviders[tip] = make(map[string]CustomProvideFunction)
	}

	inj.customProviders[tip][key] = fn
}

// GetDirectValue returns the directly provided dependency
func (inj *injector) GetDirectValue(tip reflect.Type, key string) (interface{}, bool) {
	// TODO: Avoid nil
	val, ok := inj.values[tip][key]
	return val, ok
}

// GetCustomValue returns the CustomProvideFunction for the requested dependency
func (inj *injector) GetCustomValue(tip reflect.Type, key string) (CustomProvideFunction, bool) {
	val, ok := inj.customProviders[tip][key]
	return val, ok
}

func (inj *injector) GetUnsafeValue(key string) (reflect.Value, bool) {
	val, ok := inj.unsafeValues[key]
	return val, ok
}

// CustomProvideFunction is called whenever a value is needed to be provided
// with custom logic
type CustomProvideFunction func(c *Context) (interface{}, error)
