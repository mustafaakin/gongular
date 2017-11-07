package gongular

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Engine is the main gongular router wrapper.
type Engine struct {
	// The underlying router
	actualRouter *httprouter.Router

	// Injector
	injector *injector

	// HTTP Router
	httpRouter *Router
	// WS Router
	wsRouter *WSRouter

	// The callback for route callbacks
	callback RouteCallback

	// The error handler
	errorHandler ErrorHandler
}

// NewEngine creates a new engine with the proper fields initialized
func NewEngine() *Engine {
	e := &Engine{
		errorHandler: defaultErrorHandler,
		actualRouter: httprouter.New(),
		injector:     newInjector(),
		callback:     DefaultRouteCallback,
	}

	e.httpRouter = newRouter(e)
	e.wsRouter = newWSRouter(e)
	return e
}

// GetRouter returns the underylying HTTP request router
func (e *Engine) GetRouter() *Router {
	return e.httpRouter
}

// GetWSRouter return the underlying Websocket router
func (e *Engine) GetWSRouter() *WSRouter {
	return e.wsRouter
}

// ServeFiles serves the static files
func (e *Engine) ServeFiles(path string, root http.FileSystem) {
	e.actualRouter.ServeFiles(path+"/*filepath", root)
}

// ServeFiles serves the static files
func (e *Engine) ServeFile(path, file string) {
	e.actualRouter.GET(path, func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		http.ServeFile(writer, request, file)
	})
}

// ServeHTTP serves from http
func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	e.actualRouter.ServeHTTP(w, req)
}

// GetHandler returns the underlying router as a http.Handler so that others can embed it if needed, which is
// useful for tests in our case.
func (e *Engine) GetHandler() http.Handler {
	return e.actualRouter
}

// ListenAndServe serves the given engine with a specific address. Mainly used for quick testing.
func (e *Engine) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, e.actualRouter)
}

// Provide provides with "default" key
func (e *Engine) Provide(value interface{}) {
	e.injector.Provide(value, "default")
}

// ProvideUnsafe provides a key with an exact value
func (e *Engine) ProvideUnsafe(key string, value interface{}) {
	e.injector.ProvideUnsafe(key, value)
}

// ProvideWithKey provides an interface with a key
func (e *Engine) ProvideWithKey(key string, value interface{}) {
	e.injector.Provide(value, key)
}

// CustomProvide provides with "default" key by calling the supplied CustomProvideFunction each time
func (e *Engine) CustomProvide(value interface{}, fn CustomProvideFunction) {
	e.injector.ProvideCustom(value, fn, "default")
}

// CustomProvide provides with "default" key by calling the supplied CustomProvideFunction each time
func (e *Engine) CustomProvideWithKey(key string, value interface{}, fn CustomProvideFunction) {
	e.injector.ProvideCustom(value, fn, key)
}

// SetErrorHandler sets the error handler
func (e *Engine) SetErrorHandler(fn ErrorHandler) {
	if fn == nil {
		log.Fatal("The error handler cannot be nil")
	}
	e.errorHandler = fn
}

// SetRouteCallback sets the callback function that is called when the route ends, which contains stats about the
// executed functions in that request
func (e *Engine) SetRouteCallback(fn RouteCallback) {
	e.callback = fn
}
