package gongular

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"errors"
	log "github.com/Sirupsen/logrus"
	"github.com/asaskevich/govalidator"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
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

// execute responds to an http request by using writer and request
// returns all the possible values
func (hc *handlerContext) execute(injector *Injector, w http.ResponseWriter, r *http.Request) (int, bool, interface{}, error) {
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

	// Try to fill path params such as /user/{UserId}
	if hc.param != nil {
		vars := mux.Vars(r)
		v := reflect.New(hc.param.obj).Elem()
		fields := hc.param.obj.NumField()
		for i := 0; i < fields; i++ {
			field := hc.param.obj.Field(i)
			content, ok := vars[field.Name]
			if !ok {
				validationError = fmt.Sprintf("param ", field.Name, " does not exist")
				goto fail
			} else {
				field2 := v.FieldByName(field.Name)
				kind := field2.Kind()
				if kind == reflect.Int {
					i, err := strconv.ParseInt(content, 10, 64)
					if err != nil {
						validationError = fmt.Sprintf("Expected integer for param field %s, but found '%s' instead", field.Name, content)
						goto fail
					}
					field2.SetInt(i)
				} else if kind == reflect.String {
					field2.SetString(content)
				} else {
					validationError = fmt.Sprintf("Unknown type for param field:" + content)
					goto fail
				}
			}
		}
		isValid, err := govalidator.ValidateStruct(v.Interface())
		if !isValid {
			validationError = fmt.Sprintf("Params are not valid: %s", err.Error())
			goto fail
		}
		ins[hc.param.idx] = v
	}

	// Try to parse json body
	if hc.body != nil {
		// TODO: Check type and parse accordingly, i.e. require application/json
		v := reflect.New(hc.body.obj)
		if r.Body == nil {
			validationError = "No JSON body is suppiled"
			goto fail
		}
		err := json.NewDecoder(r.Body).Decode(v.Interface())
		if err != nil {
			validationError = "Submitted request body is not JSON"
			goto fail
		}
		isValid, err := govalidator.ValidateStruct(v.Interface())
		if !isValid {
			validationError = fmt.Sprintf("Submitted body is not valid: %s", err.Error())
			goto fail
		}
		ins[hc.body.idx] = v.Elem()
	}

	if hc.query != nil {
		v := reflect.New(hc.query.obj).Elem()
		fields := hc.query.obj.NumField()

		for i := 0; i < fields; i++ {
			field := hc.query.obj.Field(i)
			content := r.URL.Query().Get(field.Name)
			if content == "" {
				validationError = fmt.Sprintf("Required query parameter not found: %s", field.Name)
				goto fail
			} else {
				// TODO: Convert it to appropriate type later
				field2 := v.FieldByName(field.Name)
				kind := field2.Kind()

				if kind == reflect.Int {
					i, err := strconv.ParseInt(content, 10, 64)
					if err != nil {
						validationError = fmt.Sprintf("Expected integer for field %s, but found '%s' instead", err.Error(), content)
						goto fail
					}
					field2.SetInt(i)
				} else if kind == reflect.String {
					field2.SetString(content)
				} else {
					validationError = fmt.Sprintf("Unknown type for field: %s", content)
					goto fail
				}
			}
		}

		isValid, err := govalidator.ValidateStruct(v.Interface())
		if !isValid {
			validationError = fmt.Sprintf("Query parameter is not valid: %s", err.Error())
			goto fail
		}
		ins[hc.query.idx] = v
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
				log.WithField("type", arg.obj).WithError(err_internal).Error("Could not provide custom value due to an internal error")
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
	if hc.outErr != nil {
		out := outs[hc.outErr.idx]
		if !out.IsNil() {
			err = out.Interface().(error)
		}
	} else if hc.outCode != nil {
		out := outs[hc.outCode.idx]
		resCode = out.Interface().(int)
	} else if hc.outResponse != nil {
		out := outs[hc.outResponse.idx]
		if !out.CanAddr() || !out.IsNil() {
			response = out.Interface()
		}
	} else if hc.outStopChain != nil {
		// bool cannot be nil (it is not *bool)
		response = outs[hc.outStopChain.idx].Bool()
	}
	// what to do with them is responsbility of the other functions
	return resCode, stopChain, response, err
}

func wrapHandlers(injector *Injector, path string, fns ...interface{}) http.HandlerFunc {
	// Determine parameter types
	hcs := make([]*handlerContext, len(fns))

	for idx, fn := range fns {
		hcs[idx] = convertHandler(injector, fn)
	}

	fn := func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		reqIdentificationNo := uuid.NewV4()

		var err2 error
		defer func() {
			r := recover()
			if r != nil {
				switch t := r.(type) {
				case string:
					err2 = errors.New(t)
				case error:
					err2 = t
				default:
					err2 = errors.New("Unknown error")
				}
				log.WithError(err2).WithField("uuid", reqIdentificationNo).Error("An error occcured while serving request")
				http.Error(w, "An internal error has occured.", http.StatusInternalServerError)
			}
		}()

		for idx, hc := range hcs {
			handlerStartTime := time.Now()
			status, stopChain, res, err := hc.execute(injector, w, r)

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "application/json")
				break
			}

			// Status is default -1
			if hc.outCode != nil || status >= 100 {
				w.WriteHeader(status)
			}

			if hc.outResponse != nil {
				w.Header().Set("Content-Type", "application/json")
				if res != nil {
					// If empty, don't return anything
					json.NewEncoder(w).Encode(res)
				}
			}

			// We stop chain if it is required, after setting status and output
			if hc.outStopChain != nil && stopChain == true {
				// Stopping the chain of execution
				break
			}

			log.WithFields(log.Fields{
				"matchedPath": path,
				"path":        r.URL.Path,
				"name":        hc.fn.String(),
				"funcId":      hc.fn,
				"idx":         idx,
				"err":         err,
				"status":      status,
				"stopChain":   stopChain,
				"res":         res,
				"elapsed":     time.Since(handlerStartTime),
				"uuid":        reqIdentificationNo,
			}).Info("HTTP Handler")
		}

		log.WithFields(log.Fields{
			"elapsed":    time.Since(startTime),
			"path":       r.URL.Path,
			"matcedPath": path,
			"uuid":       reqIdentificationNo,
		}).Info("HTTP Request Handled")
	}
	return fn
}