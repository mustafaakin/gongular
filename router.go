package gongular

import (
	"net/http"
	"path"
	"reflect"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

// Router holds information about overall router and inner objects such as
// prefix and additional handlers
type Router struct {
	muxer    *mux.Router
	injector *Injector
	prefix   string
	handlers []interface{}
}

// NewRouter initiates a router object with default params
func NewRouter() *Router {
	r := &Router{
		muxer:    mux.NewRouter(),
		injector: NewInjector(),
		prefix:   "",
		handlers: make([]interface{}, 0),
	}
	return r
}

func (r *Router) GetHandler() (http.Handler){
	return r.muxer
}

// ListenAndServe starts a web server at given addr
func (r *Router) ListenAndServe(addr string) (error) {
	log.WithField("address", addr).Info("Listening on HTTP")
	return http.ListenAndServe(addr, r.muxer)
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

	r.muxer.HandleFunc(resultingPath, wrapHandlers(r.injector, resultingPath, combinedHandlers...)).Methods("GET")
	printBindingMessage(resultingPath, "GET", combinedHandlers...)
}

// POST registers given set of handlers to a POST request at path
func (r *Router) POST(path string, handlers ...interface{}) {
	r.muxer.HandleFunc(path, wrapHandlers(r.injector, path, handlers...)).Methods("POST")
	printBindingMessage(path, "POST", handlers...)
}

// Group groups a given path with additional interfaces. It is useful to avoid
// repetitions while defining many paths
func (r *Router) Group(_path string, handlers ...interface{}) *Router {
	newRouter := &Router{
		muxer:    r.muxer,
		injector: r.injector,
		prefix:   path.Join(r.prefix, _path),
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
func (r *Router) Static(base string){
	r.muxer.PathPrefix("/").Handler(http.FileServer(http.Dir(base)))
}


// Prints the binding message for a route
func printBindingMessage(path, method string, handlers ...interface{}) {
	for idx, handler := range handlers {
		log.WithFields(log.Fields{
			"path":    path,
			"handler": reflect.TypeOf(handler),
			"method":  method,
			"index":   idx,
		}).Info("Handler registerging")
	}
}
