package gongular

import (
	"bytes"
	"log"
	"net/http"

	"path"

	"time"

	"github.com/julienschmidt/httprouter"
)

// Router holds the required states and does the mapping of requests
type Router struct {
	engine *Engine

	prefix   string
	handlers []RequestHandler
}

// NewRouter creates a new gongular2 Router
func newRouter(e *Engine) *Router {
	r := Router{
		engine:   e,
		prefix:   "",
		handlers: make([]RequestHandler, 0),
	}
	return &r
}

// GET registers the given handlers at the path for a GET request
func (r *Router) GET(path string, handlers ...RequestHandler) {
	r.combineAndWrapHandlers(path, http.MethodGet, handlers)
}

// POST registers the given handlers at the path for a POST request
func (r *Router) POST(path string, handlers ...RequestHandler) {
	r.combineAndWrapHandlers(path, http.MethodPost, handlers)
}

// PUT registers the given handlers at the path for a PUT request
func (r *Router) PUT(path string, handlers ...RequestHandler) {
	r.combineAndWrapHandlers(path, http.MethodPut, handlers)
}

// HEAD registers the given handlers at the path for a HEAD request
func (r *Router) HEAD(path string, handlers ...RequestHandler) {
	r.combineAndWrapHandlers(path, http.MethodHead, handlers)
}

// Group groups a given path with additional interfaces. It is useful to avoid
// repetitions while defining many paths
func (r *Router) Group(_path string, handlers ...RequestHandler) *Router {
	newRouter := &Router{
		engine: r.engine,
		prefix: path.Join(r.prefix, _path),
	}

	// Copy previous handlers references
	newRouter.handlers = make([]RequestHandler, len(r.handlers))
	copy(newRouter.handlers, r.handlers)

	// Append new handlers
	newRouter.handlers = append(newRouter.handlers, handlers...)

	return newRouter
}

// subpath initiates a new route with path and handlers, useful for grouping
func (r *Router) subpath(_path string, handlers []RequestHandler) (string, []RequestHandler) {
	combinedHandlers := r.handlers
	combinedHandlers = append(combinedHandlers, handlers...)

	resultingPath := path.Join(r.prefix, _path)
	return resultingPath, combinedHandlers
}

func (r *Router) combineAndWrapHandlers(path, method string, handlers []RequestHandler) {
	resultingPath, combinedHandlers := r.subpath(path, handlers)

	fn := r.transformRequestHandlers(resultingPath, method, combinedHandlers)

	switch method {
	case http.MethodGet:
		r.engine.actualRouter.GET(resultingPath, fn)
	case http.MethodPost:
		r.engine.actualRouter.POST(resultingPath, fn)
	case http.MethodPut:
		r.engine.actualRouter.PUT(resultingPath, fn)
	case http.MethodHead:
		r.engine.actualRouter.HEAD(resultingPath, fn)
	}
}

func (r *Router) transformRequestHandlers(path string, method string, handlers []RequestHandler) httprouter.Handle {
	middleHandlers := make([]*handlerContext, len(handlers))

	for i, handler := range handlers {
		mh, err := transformRequestHandler(path, method, r.engine.injector, handler)
		if err != nil {
			log.Fatal(err)
		}
		middleHandlers[i] = mh
	}

	fn := func(wr http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		st := time.Now()
		routeStat := RouteStat{
			Request:     req,
			MatchedPath: path,
			Handlers:    make([]HandlerStat, len(middleHandlers)),
		}

		// Create a logger for each request so that we can group the output
		buf := new(bytes.Buffer)
		logger := log.New(buf, "", log.LstdFlags)

		// Create a context that wraps the request, writer and logger
		ctx := contextFromRequest(path, wr, req, ps, logger)

		// For each of the handler this route has, try to execute it
		for idx, handler := range middleHandlers {
			hc := HandlerStat{
				FuncName: handler.name,
			}

			// Parse the parameters to the handler object
			stHandler := time.Now()
			fn := handler.RequestHandler
			err := fn(ctx)

			hc.Duration = time.Since(stHandler)

			// If an error occurs, stop the chain
			if err != nil {
				ctx.StopChain()
				r.engine.errorHandler(err, ctx)

				// Put the route stats
				hc.Error = err
				hc.StopChain = true
				routeStat.Handlers[idx] = hc

				break
			}

			// Voluntarily stopped
			if ctx.stopChain {
				// Put the route stats
				hc.Duration = time.Since(st)
				hc.StopChain = true
				routeStat.Handlers[idx] = hc

				break
			}

			routeStat.Handlers[idx] = hc
		}

		// Save final stats
		routeStat.ResponseSize = ctx.Finalize()
		routeStat.ResponseCode = ctx.status
		routeStat.TotalDuration = time.Since(st)
		routeStat.Logs = buf

		if r.engine.callback != nil {
			r.engine.callback(routeStat)
		}
	}

	return fn
}
