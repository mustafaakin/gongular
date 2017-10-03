package gongular

import (
	"bytes"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
)

// WebsocketHandler handles the Websocket interactions. It has two functions Before, Handle which must be implemented
// by the target object
type WebsocketHandler interface {
	// Before is a filter applied just before upgrading the request to websocket. It can be useful for filtering the
	// request and returning an error would not open a websocket but close it with an error. The http.Header is for
	// answering with a http.Header which allows setting a cookie. Can be omitted if not desired.
	Before(c *Context) (http.Header, error)
	// Handle is regular handling of the web socket, user is fully responsible for the request
	Handle(conn *websocket.Conn)
}

// WSRouter wraps the Engine with the ability to map WebsocketHandler to routes. Currently it does not support
// sub-routing but it supports injections, param and query parameter bindings.
type WSRouter struct {
	engine *Engine
}

func newWSRouter(e *Engine) *WSRouter {
	return &WSRouter{
		engine: e,
	}
}

// Handle registers the given Websocket handler if
func (r *WSRouter) Handle(path string, handler WebsocketHandler) {
	mh, err := transformWebsocketHandler(path, r.engine.injector, handler)
	if err != nil {
		log.Fatal(err)
	}

	fn := func(wr http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		// Create a logger for each request so that we can group the output
		buf := new(bytes.Buffer)
		logger := log.New(buf, "", log.LstdFlags)

		// Create a context that wraps the request, writer and logger
		ctx := contextFromRequest(path, wr, req, ps, logger)

		// Parse the parameters to the handler object
		fn := mh.RequestHandler
		err := fn(ctx)
		if err != nil {
			r.engine.errorHandler(err, ctx)
		}
	}

	r.engine.actualRouter.GET(path, fn)
}
