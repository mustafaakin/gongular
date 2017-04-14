package gongular

import (
	"net/http"
	"path"
	"reflect"

	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/julienschmidt/httprouter"
)

// ErrorHandle is the function signature that must be implemented to provide a custom Error Handler
type ErrorHandle func(e error, c *Context)
type PanicHandle func(v interface{}, c *Context)

// DefaultErrorHandle writes 500 and displays error to user.
var DefaultErrorHandle = func(err error, c *Context) {
	c.StopChain()
	c.MustStatus(http.StatusInternalServerError)
	c.SetBody(map[string]string{
		"error": err.Error(),
	})
}

var DefaultPanicHandle = func(v interface{}, c *Context) {
	c.StopChain()
	c.MustStatus(http.StatusInternalServerError)
	c.SetBody(map[string]interface{}{
		"panic": v,
	})
	c.Finalize()
}

// Router holds information about overall router and inner objects such as
// prefix and additional handlers
type Router struct {
	router       *httprouter.Router
	injector     *Injector
	prefix       string
	handlers     []interface{}
	InfoLog      *log.Logger
	DebugLog     *log.Logger
	errorHandler ErrorHandle
}

// NewRouter initiates a router object with default params
func NewRouter() *Router {
	r := &Router{
		router:       httprouter.New(),
		injector:     NewInjector(),
		prefix:       "",
		handlers:     make([]interface{}, 0),
		InfoLog:      log.New(os.Stdout, "[INFO ] ", log.LstdFlags),
		DebugLog:     log.New(os.Stdout, "[DEBUG] ", log.LstdFlags),
		errorHandler: DefaultErrorHandle,
	}

	r.SetPanicHandler(DefaultPanicHandle)

	return r
}

// NewRouterTest initiates a router object with default params
func NewRouterTest() *Router {
	r := NewRouter()
	r.DebugLog.SetOutput(ioutil.Discard)
	r.DebugLog.SetFlags(0)
	r.InfoLog.SetOutput(ioutil.Discard)
	r.InfoLog.SetFlags(0)
	return r
}

//
func (r *Router) SetPanicHandler(fn PanicHandle) {
	r.router.PanicHandler = func(w http.ResponseWriter, req *http.Request, v interface{}) {
		c := ContextFromRequest(w, req, r.InfoLog)
		fn(v, c)
	}
}

func (r *Router) SetErrorHandler(fn ErrorHandle) {
	r.errorHandler = fn
}

// DisableDebug disables the debug outputs that might be too much for some people
func (r *Router) DisableDebug() {
	r.DebugLog.SetOutput(ioutil.Discard)
	r.DebugLog.SetFlags(0)
}

// EnableDebug enables the debug outputs
func (r *Router) EnableDebug() {
	r.DebugLog.SetOutput(os.Stdout)
	r.DebugLog.SetFlags(log.LstdFlags)
}

// GetHandler returns the http.Handler so that it can be used in HTTP servers
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
	combinedHandlers = append(combinedHandlers, handlers...)

	resultingPath := path.Join(r.prefix, _path)
	return resultingPath, combinedHandlers
}

func (r *Router) combineAndWrapHandlers(path, method string, handlers ...interface{}) {
	resultingPath, combinedHandlers := r.subpath(path, handlers)
	fn := r.wrapHandlers(r.injector, resultingPath, combinedHandlers...)
	r.printBindingMessage(resultingPath, method, combinedHandlers...)

	if method == "GET" {
		r.router.GET(resultingPath, fn)
	} else if method == "POST" {
		r.router.POST(resultingPath, fn)
	}
}

// GET registers given set of handlers to a GET request at path
func (r *Router) GET(_path string, handlers ...interface{}) {
	r.combineAndWrapHandlers(_path, http.MethodGet, handlers...)
}

// POST registers given set of handlers to a POST request at path
func (r *Router) POST(_path string, handlers ...interface{}) {
	r.combineAndWrapHandlers(_path, http.MethodPost, handlers...)
}

// Group groups a given path with additional interfaces. It is useful to avoid
// repetitions while defining many paths
func (r *Router) Group(_path string, handlers ...interface{}) *Router {
	newRouter := &Router{
		router:       r.router,
		injector:     r.injector,
		prefix:       path.Join(r.prefix, _path),
		InfoLog:      r.InfoLog,
		DebugLog:     r.DebugLog,
		errorHandler: r.errorHandler,
	}

	// Copy previous handlers references
	newRouter.handlers = make([]interface{}, len(r.handlers))
	copy(newRouter.handlers, r.handlers)

	// Append new handlers
	newRouter.handlers = append(newRouter.handlers, handlers...)

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
		r.InfoLog.Printf("%-5s %-40s %-20s\n", method, path, reflect.TypeOf(handler))
	}
}

func (r *Router) wrapHandlers(injector *Injector, path string, fns ...interface{}) httprouter.Handle {
	// Determine parameter types
	hcs := make([]*handlerContext, len(fns))
	for idx, fn := range fns {
		hcs[idx] = convertHandler(injector, fn)
	}

	fn := func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		startTime := time.Now()

		// Create a context that will be used among multiple headers
		c := ContextFromRequest(w, req, r.InfoLog)

		// TODO: Eliminate this with using custom error handler of httprouter
		/*
			defer func() {
				rec := recover()
				if rec != nil {
					var err error
					switch t := rec.(type) {
					case string:
						err = errors.New(t)
					case error:
						err = t
					default:
						err = errors.New("Unknown error")
					}

					r.InfoLog.Println("A panic occured while serving request: " + err.Error())
					r.ErrorHandler(err, c)
				}
			}()
		*/

		for _, hc := range hcs {
			handlerStartTime := time.Now()
			res, err := hc.execute(injector, c, ps)

			if err != nil {
				r.errorHandler(err, c)
				break // Stopping the chain
			}

			// We stop chain if it is required
			if c.stopChain {
				break
			}

			// If error is nil, and user is returning error, its his problem
			if hc.outResponse != nil {
				if res != nil {
					c.SetBody(res)
				}
				// TODO: Else what?
			}

			r.DebugLog.Printf("%-5s %-30s %-30s %10s\n", req.Method, req.URL.Path, hc.fn.String(), time.Since(handlerStartTime).String())
		}

		// Finally write the request to client
		bytes := c.Finalize()
		r.InfoLog.Printf("%-5s %-30s %-30s %10s %4d %d\n", req.Method, req.URL.Path, path, time.Since(startTime).String(), c.status, bytes)
	}

	return fn
}
