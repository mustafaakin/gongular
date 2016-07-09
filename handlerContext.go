package gongular

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"errors"
	"github.com/asaskevich/govalidator"
	"github.com/julienschmidt/httprouter"
	"log"
)

var ErrNoJsonBody = errors.New("No JSON body is suppiled")
var ErrNotValidJsonBody = errors.New("Submitted request body is not JSON")

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
	resWriter *funcArg
	req       *funcArg

	// Input Fields
	query *funcArg
	body  *funcArg
	form  *funcArg
	param *funcArg

	// Output Fields
	outErr       *funcArg
	outResponse  *funcArg
	outCode      *funcArg
	outStopChain *funcArg

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
		panic("Wrapped interface is not a function.")
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

		// Look if it is in supplied version
		if _, ok := ij.values[in]; ok {
			hc.args = append(hc.args, &funcArg{
				idx: i,
				obj: in,
			})
		} else if _, ok := ij.customProviders[in]; ok {
			hc.customArgs = append(hc.customArgs, &funcArg{
				idx: i,
				obj: in,
			})
		} else {
			httpWriterType := reflect.TypeOf((*http.ResponseWriter)(nil)).Elem()
			if in.Implements(httpWriterType) {
				hc.resWriter = &funcArg{
					idx: i,
					obj: in,
				}
			} else if in.AssignableTo(reflect.TypeOf(&http.Request{})) {
				hc.req = &funcArg{
					idx: i,
					obj: in,
				}
			} else {
				// Get its name and see if it ends with Query, Body, Form or Param
				name := in.String()
				arg := &funcArg{
					idx: i,
					obj: in,
				}
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

		if t == reflect.Bool {
			// Bool is used in middleware to stop the chain.
			// False is used because if it did not return true, it means abort
			hc.outStopChain = arg
		} else if t == reflect.Int {
			// Int has precedence because it is expected to indicate the response status code
			hc.outCode = arg
		} else if t == reflect.Interface {
			// Checks if error, see: http://stackoverflow.com/questions/30688514/go-reflect-how-to-check-whether-reflect-type-is-an-error-type
			errType := reflect.TypeOf((*error)(nil)).Elem()
			if out.Implements(errType) {
				hc.outErr = arg
			} else {
				hc.outResponse = arg
			}
		} else {
			hc.outResponse = arg
		}
	}

	return hc
}

func (hc *handlerContext) parseParams(ps httprouter.Params) (*reflect.Value, error) {
	v := reflect.New(hc.param.obj).Elem()
	fields := hc.param.obj.NumField()
	for i := 0; i < fields; i++ {
		field := hc.param.obj.Field(i)
		content := ps.ByName(field.Name)
		if content == "" {
			validationError := fmt.Sprintf("param ", field.Name, " does not exist")
			return nil, errors.New(validationError)
		} else {
			field2 := v.FieldByName(field.Name)
			kind := field2.Kind()
			if kind == reflect.Int {
				i, err := strconv.ParseInt(content, 10, 64)
				if err != nil {
					validationError := fmt.Sprintf("Expected integer for param field %s, but found '%s' instead", field.Name, content)
					return nil, errors.New(validationError)
				}
				field2.SetInt(i)
			} else if kind == reflect.String {
				field2.SetString(content)
			} else {
				validationError := fmt.Sprintf("Unknown type for param field:" + content)
				return nil, errors.New(validationError)
			}
		}
	}

	isValid, err := govalidator.ValidateStruct(v.Interface())
	if !isValid {
		validationError := fmt.Sprintf("Params are not valid: %s", err.Error())
		return nil, errors.New(validationError)
	}

	return &v, nil
}

func (hc *handlerContext) parseBody(r *http.Request) (*reflect.Value, error) {
	// Check if body exists so we try to parse it
	if r.Body == nil {
		return nil, ErrNoJsonBody
	}

	// Construct given object
	v := reflect.New(hc.body.obj)

	// Try to parse it to our interface
	err := json.NewDecoder(r.Body).Decode(v.Interface())
	if err != nil {
		return nil, ErrNotValidJsonBody
	}

	// When parsing done, validate it
	isValid, err := govalidator.ValidateStruct(v.Interface())
	if !isValid {
		validationError := fmt.Sprintf("Submitted body is not valid: %s", err.Error())
		return nil, errors.New(validationError)
	}

	// Return the final element
	elem := v.Elem()
	return &elem, nil
}

func (hc *handlerContext) parseQuery(r *http.Request) (*reflect.Value, error){
	v := reflect.New(hc.query.obj).Elem()
	fields := hc.query.obj.NumField()

	for i := 0; i < fields; i++ {
		field := hc.query.obj.Field(i)
		content := r.URL.Query().Get(field.Name)
		if content == "" {
			validationError := fmt.Sprintf("Required query parameter not found: %s", field.Name)
			return nil, errors.New(validationError)
		} else {
			// TODO: Convert it to appropriate type later
			field2 := v.FieldByName(field.Name)
			kind := field2.Kind()

			if kind == reflect.Int {
				i, err := strconv.ParseInt(content, 10, 64)
				if err != nil {
					validationError := fmt.Sprintf("Expected integer for field %s, but found '%s' instead", err.Error(), content)
					return nil, errors.New(validationError)
				}
				field2.SetInt(i)
			} else if kind == reflect.String {
				field2.SetString(content)
			} else {
				validationError := fmt.Sprintf("Unknown type for field: %s", content)
				return nil, errors.New(validationError)
			}
		}
	}

	isValid, err := govalidator.ValidateStruct(v.Interface())
	if !isValid {
		validationError := fmt.Sprintf("Query parameter is not valid: %s", err.Error())
		return nil, errors.New(validationError)
	}

	return &v, nil
}

// execute responds to an http request by using writer and request
// returns all the possible values
func (hc *handlerContext) execute(injector *Injector, w http.ResponseWriter, r *http.Request, ps httprouter.Params, logger *log.Logger) (int, bool, interface{}, error) {
	// Prepare inputs to be supplied to hc.fn function
	ins := make([]reflect.Value, hc.numIn)

	validationError := ""

	// Just put http.ResponseWriter as is
	if hc.resWriter != nil {
		ins[hc.resWriter.idx] = reflect.ValueOf(w)
	}

	// Just put the *http.Request as is
	if hc.req != nil {
		ins[hc.req.idx] = reflect.ValueOf(r)
	}

	// Try to fill path params such as /user/:UserId
	if hc.param != nil {
		v, err := hc.parseParams(ps)
		if err == nil {
			ins[hc.param.idx] = *v
		} else {
			goto fail
		}
	}

	// Try to parse json body
	if hc.body != nil {
		// TODO: Check type and parse accordingly, i.e. require application/json
		v, err := hc.parseBody(r)
		if err != nil {
			ins[hc.body.idx] = *v
		} else {
			goto fail
		}
	}

	if hc.query != nil {
		v, err := hc.parseQuery(r)
		if err != nil {
			ins[hc.query.idx] = *v
		} else {
			goto fail
		}
	}

	// Try to put as-is dependencies such as db connections
	for _, arg := range hc.args {
		// Check if it exists on just value dependencies first
		if val, ok := injector.values[arg.obj]; ok {
			ins[arg.idx] = reflect.ValueOf(val)
		} else {
			panic("Dont know how to inject!")
		}
	}

	// Try to put custom provided dependencies such as custom logic that might
	// be required to get user info from session
	for _, arg := range hc.customArgs {
		// Check if it exists on execution-injectable values then
		if fn, ok := injector.customProviders[arg.obj]; ok {
			err_internal, out := fn(w, r)
			if err_internal != nil {
				logger.Printf("Could not provide custom value '%s' to do an error: '%s'\n", arg.obj, err_internal)
				return http.StatusInternalServerError, true, "An internal error has occured", nil
			} else if out == nil {
				return -1, true, "", nil
			} else {
				ins[arg.idx] = reflect.ValueOf(out)
			}
		} else {
			panic("Don't know how to inject!")
		}
	}

	goto nofail
fail:
	return http.StatusBadRequest, true, validationError, nil
nofail:
	// Call the function with supplied values
	outs := hc.fn.Call(ins)

	var err error
	resCode := -1
	var stopChain bool
	var response interface{}

	// TODO: Wrong logic here fix, why else if? we set them no matter what
	if hc.outErr != nil {
		out := outs[hc.outErr.idx]
		if !out.IsNil() {
			err = out.Interface().(error)
		}
	}

	if hc.outCode != nil {
		out := outs[hc.outCode.idx]
		resCode = out.Interface().(int)
	}

	if hc.outResponse != nil {
		out := outs[hc.outResponse.idx]
		// else?
		if !out.CanAddr() || !out.IsNil() {
			response = out.Interface()
		}
	}

	if hc.outStopChain != nil {
		// bool cannot be nil (it is not *bool)
		response = outs[hc.outStopChain.idx].Bool()
	}

	// what to do with them is responsibility of the other functions
	return resCode, stopChain, response, err
}