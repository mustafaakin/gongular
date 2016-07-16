package gongular

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/asaskevich/govalidator"
	"github.com/julienschmidt/httprouter"
)

// funcArg remembers the type of the function argument and its index
type funcArg struct {
	obj reflect.Type
	idx int
}

// handlerContext stores all the information to respond to an HTTP request
type handlerContext struct {
	// Injected Dependencies
	args       []*funcArg
	customArgs []*funcArg

	// HTTP Request , Response Writer
	context *funcArg

	// Input Fields
	query *funcArg
	body  *funcArg
	form  *funcArg
	param *funcArg

	// Output Fields
	outErr      *funcArg
	outResponse *funcArg

	// Function information
	fn     reflect.Value
	numIn  int
	numOut int
}

// Given an injector and function, it creates a hanlderContext to respond
// to http requests
func convertHandler(ij *Injector, fn interface{}) *handlerContext {
	t := reflect.TypeOf(fn)
	if t.Kind() != reflect.Func {
		panic("Wrapped interface is not a function, it is a " + t.Kind().String())
	}

	// Preserve the information about the function here so that
	// we do not have to use reflection more than neeeded in each request
	hc := &handlerContext{
		args:       make([]*funcArg, 0),
		customArgs: make([]*funcArg, 0),
	}

	// Analyze the input parameters first
	for i := 0; i < t.NumIn(); i++ {
		in := t.In(i)

		arg := &funcArg{
			idx: i,
			obj: in,
		}

		// Look if it is in supplied version
		if _, ok := ij.values[in]; ok {
			hc.args = append(hc.args, arg)
		} else if _, ok := ij.customProviders[in]; ok {
			hc.customArgs = append(hc.customArgs, arg)
		} else if in.AssignableTo(reflect.TypeOf(&Context{})) {
			hc.context = arg
		} else {
			// Get its name and see if it ends with Query, Body, Form or Param
			name := in.String()
			if strings.HasSuffix(name, "Body") {
				hc.body = arg
			} else if strings.HasSuffix(name, "Form") {
				hc.form = arg
			} else if strings.HasSuffix(name, "Query") {
				hc.query = arg
			} else if strings.HasSuffix(name, "Param") {
				hc.param = arg
			} else {
				panic("Unknown parameter:" + fmt.Sprintf("%s %s", fn, in))
			}
		}
	}

	// Remember the function
	hc.fn = reflect.ValueOf(fn)
	hc.numIn = t.NumIn()
	hc.numOut = t.NumOut()

	for i := 0; i < t.NumOut(); i++ {
		out := t.Out(i)
		t := out.Kind()

		arg := &funcArg{
			idx: i,
			obj: out,
		}

		if t == reflect.Struct {
			hc.outResponse = arg
		} else if t == reflect.Interface {
			// Checks if error, see: http://stackoverflow.com/questions/30688514/go-reflect-how-to-check-whether-reflect-type-is-an-error-type
			errType := reflect.TypeOf((*error)(nil)).Elem()
			if out.Implements(errType) {
				hc.outErr = arg
			}
		} else {
			hc.outResponse = arg
		}
	}

	return hc
}

func (hc *handlerContext) parseParams(ps httprouter.Params) (*reflect.Value, string) {
	v := reflect.New(hc.param.obj).Elem()
	fields := hc.param.obj.NumField()
	for i := 0; i < fields; i++ {
		field := hc.param.obj.Field(i)
		content := ps.ByName(field.Name)

		field2 := v.FieldByName(field.Name)
		kind := field2.Kind()
		if kind == reflect.Int {
			i, err := strconv.ParseInt(content, 10, 64)
			if err != nil {
				return nil, fmt.Sprintf("Expected integer for param field %s, but found '%s' instead", field.Name, content)
			}
			field2.SetInt(i)
		} else if kind == reflect.String {
			field2.SetString(content)
		} else {
			return nil, fmt.Sprintf("Unknown type for param field:" + content)
		}

	}

	isValid, err := govalidator.ValidateStruct(v.Interface())
	if !isValid {
		return nil, fmt.Sprintf("Params are not valid: %s", err.Error())
	}

	return &v, ""
}

func (hc *handlerContext) parseBody(r *http.Request) (*reflect.Value, string) {
	// Check if body exists so we try to parse it
	if r.Body == nil {
		return nil, "No request body is supplied"
	}

	// Construct given object
	v := reflect.New(hc.body.obj)

	// Try to parse it to our interface
	err := json.NewDecoder(r.Body).Decode(v.Interface())
	if err != nil {
		return nil, "Supplied request body is not valid json: " + err.Error()
	}

	// When parsing done, validate it
	isValid, err := govalidator.ValidateStruct(v.Interface())
	if !isValid {
		return nil, fmt.Sprintf("Submitted body is not valid: %s", err.Error())
	}

	// Return the final element
	elem := v.Elem()
	return &elem, ""
}

func (hc *handlerContext) parseQuery(r *http.Request) (*reflect.Value, string) {
	v := reflect.New(hc.query.obj).Elem()
	fields := hc.query.obj.NumField()

	for i := 0; i < fields; i++ {
		field := hc.query.obj.Field(i)
		content := r.URL.Query().Get(field.Name)

		if content == "" {
			// TODO: Figure out what to do in absence, the concern should be validator's
			return nil, fmt.Sprintf("Required query parameter not found: %s", field.Name)
		}

		field2 := v.FieldByName(field.Name)
		kind := field2.Kind()

		if kind == reflect.Int {
			i, err := strconv.ParseInt(content, 10, 64)
			if err != nil {
				return nil, fmt.Sprintf("Expected integer for field %s, but found '%s' instead", err.Error(), content)
			}
			field2.SetInt(i)
		} else if kind == reflect.String {
			field2.SetString(content)
		} else {
			return nil, fmt.Sprintf("Unknown type for field: %s", content)
		}
	}

	isValid, err := govalidator.ValidateStruct(v.Interface())
	if !isValid {
		return nil, fmt.Sprintf("Query parameter is not valid: %s", err.Error())
	}

	return &v, ""
}

// execute responds to an http request by using writer and request
// returns all the possible values
func (hc *handlerContext) execute(injector *Injector, c *Context, ps httprouter.Params) (interface{}, error) {
	// Prepare inputs to be supplied to hc.fn function
	ins := make([]reflect.Value, hc.numIn)

	// Create a Gongular.Context object from req
	if hc.context != nil {
		ins[hc.context.idx] = reflect.ValueOf(c)
	}

	// Try to fill path params such as /user/:UserId
	if hc.param != nil {
		v, validationError := hc.parseParams(ps)
		if validationError == "" {
			ins[hc.param.idx] = *v
		} else {
			c.Fail(http.StatusBadRequest, validationError)
			return nil, nil
		}
	}

	// Try to parse json body
	if hc.body != nil {
		// TODO: Check type and parse accordingly, i.e. require application/json
		v, validationError := hc.parseBody(c.r)
		if validationError == "" {
			ins[hc.body.idx] = *v
		} else {
			c.Fail(http.StatusBadRequest, validationError)
			return nil, nil
		}
	}

	if hc.query != nil {
		v, validationError := hc.parseQuery(c.r)
		if validationError == "" {
			ins[hc.query.idx] = *v
		} else {
			c.Fail(http.StatusBadRequest, validationError)
			return nil, nil
		}
	}

	// Try to put as-is dependencies such as db connections
	for _, arg := range hc.args {
		// Check if it exists on just value dependencies first
		if val, ok := injector.values[arg.obj]; ok {
			ins[arg.idx] = reflect.ValueOf(val)
		}
	}

	// Try to put custom provided dependencies such as custom logic that might
	// be required to get user info from session
	for _, arg := range hc.customArgs {
		// Check if it exists on execution-injectable values then
		if fn, ok := injector.customProviders[arg.obj]; ok {
			errInternal, out := fn(c)
			if errInternal != nil {
				c.logger.Printf("Could not provide custom value '%s' to do an error: '%s'\n", arg.obj, errInternal)
				c.Fail(http.StatusInternalServerError, "An internal error has occured")
				return nil, nil
			} else if out == nil {
				// TODO: Nil provided? Log?
				c.StopChain()
				return nil, nil
			} else {
				ins[arg.idx] = reflect.ValueOf(out)
			}
		} else {
			panic("Don't know how to inject!")
		}
	}

	// Call the function with supplied values
	outs := hc.fn.Call(ins)

	var err error
	var response interface{}

	// TODO: Wrong logic here fix, why else if? we set them no matter what
	if hc.outErr != nil {
		out := outs[hc.outErr.idx]
		if !out.IsNil() {
			err = out.Interface().(error)
		}
	}

	if hc.outResponse != nil {
		out := outs[hc.outResponse.idx]
		// else?
		if !out.CanAddr() || !out.IsNil() {
			response = out.Interface()
		}
	}

	// what to do with them is responsibility of the other functions
	return response, err
}
