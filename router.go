package gongular

import (
	"net/http"
	"path"
	"reflect"

	"encoding/json"
	"errors"
	"github.com/julienschmidt/httprouter"
	"log"
	"time"
	"os"
	"io/ioutil"
)

// Router holds information about overall router and inner objects such as
// prefix and additional handlers
type Router struct {
	router    *httprouter.Router
	injector  *Injector
	prefix    string
	handlers  []interface{}
	InfoLog   *log.Logger
	DebugLog  *log.Logger
}

// NewRouter initiates a router object with default params
func NewRouter() *Router {
	r := &Router{
		router:    httprouter.New(),
		injector:  NewInjector(),
		prefix:    "",
		handlers:  make([]interface{}, 0),
		InfoLog :  log.New(os.Stdout, "[INFO ] ", log.LstdFlags),
		DebugLog:  log.New(os.Stdout, "[DEBUG] ", log.LstdFlags),
	}

	// TODO: Set panic handler
	// r.router.PanicHandler =
	return r
}

func (r *Router) DisableDebug(){
	r.DebugLog.SetOutput(ioutil.Discard)
	r.DebugLog.SetFlags(0)
}

func (r *Router) EnableDebug(){
	r.DebugLog.SetOutput(os.Stdout)
	r.DebugLog.SetFlags(log.LstdFlags)
}

func (r *Router) GetHandler() http.Handler {
	return r.router
}

// ListenAndServe starts a web server at given addr
func (r *Router) ListenAndServe(addr string) error {
	r.InfoLog.Println("Listening HTTP on " + addr)
	return http.ListenAndServe(addr, r.router)
}

// subpath initiates a new route with path and handlers, useful for grouping
func (r *Router) subpath(_path string, handlers []interface{}) (string, []interface{}) {
	combinedHandlers := r.handlers
	for _, handler := range handlers {
		combinedHandlers = append(combinedHandlers, handler)
	}
	resultingPath := path.Join(r.prefix, _path)
	return resultingPath, combinedHandlers
}

// GET registers given set of handlers to a GET request at path
func (r *Router) GET(_path string, handlers ...interface{}) {
	resultingPath, combinedHandlers := r.subpath(_path, handlers)

	fn := r.wrapHandlers(r.injector, resultingPath, combinedHandlers...)
	r.router.GET(resultingPath, fn)
	r.printBindingMessage(resultingPath, "GET", combinedHandlers...)
}

// POST registers given set of handlers to a POST request at path
func (r *Router) POST(_path string, handlers ...interface{}) {
	resultingPath, combinedHandlers := r.subpath(_path, handlers)

	fn := r.wrapHandlers(r.injector, resultingPath, combinedHandlers...)
	r.router.GET(resultingPath, fn)
	r.printBindingMessage(resultingPath, "POST", handlers...)
}

// Group groups a given path with additional interfaces. It is useful to avoid
// repetitions while defining many paths
func (r *Router) Group(_path string, handlers ...interface{}) *Router {
	newRouter := &Router{
		router:    r.router,
		injector:  r.injector,
		prefix:    path.Join(r.prefix, _path),
	}

	// Copy previous handlers references
	copy(r.handlers, newRouter.handlers)

	// Append new handlers
	for _, handler := range handlers {
		newRouter.handlers = append(newRouter.handlers, handler)
	}

	return newRouter
}

// Provide tells the injector to use the given value
func (r *Router) Provide(value interface{}) {
	r.injector.Provide(value)
}

// ProvideCustom tells the injector to use the given value type with given
// CustomProvideFunction
func (r *Router) ProvideCustom(value interface{}, fn CustomProvideFunction) {
	r.injector.ProvideCustom(value, fn)
}

// Static serves static files from a given base, without any prefix
func (r *Router) Static(prefix, base string) {
	r.router.ServeFiles(path.Join(prefix, "*filepath"), http.Dir(base))
}

// Prints the binding message for a route
func (r *Router) printBindingMessage(path, method string, handlers ...interface{}) {
	for _, handler := range handlers {
		r.DebugLog.Printf("%-5s %-40s %-20s\n", method, path, reflect.TypeOf(handler))
	}
}

func (router *Router) wrapHandlers(injector *Injector, path string, fns ...interface{}) httprouter.Handle {
	// Determine parameter types
	hcs := make([]*handlerContext, len(fns))
	for idx, fn := range fns {
		hcs[idx] = convertHandler(injector, fn)
	}

	fn := func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		startTime := time.Now()
		// reqIdentificationNo := uuid.NewV4()
		responseStatus := -1;

		// TODO: Eliminate this with using custom error handler of httprouter
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
				// log.WithError(err2).WithField("uuid", reqIdentificationNo).Error("An error occcured while serving request")
				http.Error(w, "An internal error has occured.", http.StatusInternalServerError)
			}
		}()

		for _, hc := range hcs {
			handlerStartTime := time.Now()
			status, stopChain, res, err := hc.execute(injector, w, r, ps, router.InfoLog)

			if err != nil {
				// TODO: Custom error handler
				responseStatus = http.StatusInternalServerError
				w.WriteHeader(responseStatus)
				w.Header().Set("Content-Type", "application/json")
				break
			}

			// Status is default -1
			if responseStatus  != -1 && hc.outCode != nil || status >= 100 {
				responseStatus = status
				w.WriteHeader(status)
			}

			if hc.outResponse != nil {
				// If empty, don't return anything
				if res != nil {
					w.Header().Set("Content-Type", "application/json")
					if responseStatus == -1 {
						// Meaning no one set the status before
						responseStatus = http.StatusOK
					}
					json.NewEncoder(w).Encode(res)
				}
				// TODO: Else what?
			}

			// We stop chain if it is required, after setting status and output
			if hc.outStopChain != nil && stopChain == true {
				// Stopping the chain of execution
				break
			}

			router.DebugLog.Printf("%-5s %-40s %-40s %5s\n", r.Method, r.URL.Path, hc.fn.String(), time.Since(handlerStartTime).String())
		}
		router.InfoLog.Printf("%-5s %-40s %-40s %5s %4d\n", r.Method, r.URL.Path, path, time.Since(startTime).String(), responseStatus)
	}

	return fn
}
