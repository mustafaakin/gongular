package gongular

import (
	"errors"
	"reflect"

	"net/http"

	"fmt"

	"github.com/gorilla/websocket"
)

// RequestHandler is a generic handler for gongular2
type RequestHandler interface {
	Handle(c *Context) error
}

type middleRequestHandler func(c *Context) error

type handlerContext struct {
	method    string
	name      string
	websocket bool
	// The analyzed reflection data so that we can cache it
	param     bool
	query     bool
	body      bool
	form      bool
	injection bool

	// HandlerType
	tip reflect.Type

	// The actual function
	RequestHandler middleRequestHandler
}

func (hc *handlerContext) checkParam(handlerElem reflect.Type) error {
	// See if we have any params
	param, paramOk := handlerElem.FieldByName(FieldParameter)
	if paramOk {
		// If we have something param, it should be a struct only
		// TODO: Additional check, it should be flat struct
		// TODO: Additional check, it should be compatible with path
		if param.Type.Kind() != reflect.Struct {
			return errors.New("Param field added but it is not a struct")
		}
	}
	hc.param = paramOk
	return nil
}

func (hc *handlerContext) checkQuery(handlerElem reflect.Type) error {
	query, queryOk := handlerElem.FieldByName(FieldQuery)
	if queryOk {
		// If we have something param, it should be a struct only
		// TODO: Additional check, it should be flat struct
		if query.Type.Kind() != reflect.Struct {
			return errors.New("Query field added but it is not a struct")
		}
	}
	hc.query = queryOk
	return nil
}

func (hc *handlerContext) checkBody(handlerElem reflect.Type) error {
	_, bodyOk := handlerElem.FieldByName(FieldBody)
	hc.body = bodyOk
	return nil
}

func (hc *handlerContext) checkForm(handlerElem reflect.Type) error {
	_, formOk := handlerElem.FieldByName(FieldForm)
	hc.form = formOk
	return nil

}

func (hc *handlerContext) checkRequestFields(handlerElem reflect.Type) error {
	err := hc.checkBody(handlerElem)
	if err != nil {
		return err
	}

	err = hc.checkForm(handlerElem)
	if err != nil {
		return err
	}

	// TODO: Add path field check to see whether path has those variables?
	err = hc.checkParam(handlerElem)
	if err != nil {
		return err
	}

	err = hc.checkQuery(handlerElem)
	return err
}

func transformRequestHandler(path string, method string, injector *injector, handler RequestHandler) (*handlerContext, error) {
	rhc := handlerContext{}
	// Handler parse parameters
	handlerElem := reflect.TypeOf(handler).Elem()
	rhc.name = fmt.Sprintf("%s.%s", handlerElem.PkgPath(), handlerElem.Name())
	rhc.tip = handlerElem
	rhc.method = method

	err := rhc.checkRequestFields(handlerElem)
	if err != nil {
		return nil, err
	}

	if method == http.MethodGet {
		if rhc.form || rhc.body {
			return nil, errors.New("A GET request handler cannot have body or form")
		}
	}

	for i := 0; i < handlerElem.NumField(); i++ {
		name := handlerElem.Field(i).Name
		if name == FieldBody || name == FieldForm || name == FieldQuery || name == FieldParameter {
			continue
		} else {
			// TODO: Check if we can set it!, is the field exported?
			rhc.injection = true
			break
		}
	}

	rhc.RequestHandler = rhc.getMiddleRequestHandler(injector)
	return &rhc, nil
}

func transformWebsocketHandler(path string, injector *injector, handler WebsocketHandler) (*handlerContext, error) {
	hc := &handlerContext{
		websocket: true,
	}

	// Handler parse parameters
	handlerElem := reflect.TypeOf(handler).Elem()

	hc.tip = handlerElem

	err := hc.checkRequestFields(handlerElem)
	if err != nil {
		return nil, err
	}

	hc.RequestHandler = hc.getMiddleRequestHandler(injector)
	return hc, nil
}

func (hc *handlerContext) parseFields(c *Context, objElem reflect.Value, injector *injector) error {
	if hc.param {
		err := c.parseParams(objElem)
		if err != nil {
			return err
		}
	}

	if hc.query {
		err := c.parseQuery(objElem)
		if err != nil {
			return err
		}
	}

	if hc.body {
		err := c.parseBody(objElem)
		if err != nil {
			return err
		}
	}

	if hc.form {
		err := c.parseForm(objElem)
		if err != nil {
			return err
		}
	}

	if hc.injection {
		err := c.parseInjections(objElem, injector)
		return err
	}
	return nil
}

func (hc *handlerContext) executeWebsocketHandler(obj reflect.Value, c *Context) error {
	wsHandler, ok := obj.Interface().(WebsocketHandler)
	if !ok {
		// It should, it cannot be here
		return errors.New("The interface does not implement WebsocketHandler: " + hc.tip.Name())
	}

	responseHeader, err := wsHandler.Before(c)
	if err != nil {
		return err
	}

	var upgrader = websocket.Upgrader{}

	conn, err := upgrader.Upgrade(c.w, c.r, responseHeader)
	if err != nil {
		return err
	}

	wsHandler.Handle(conn)
	return nil
}

func (hc *handlerContext) executeRequestHandler(obj reflect.Value, c *Context) error {
	// If it is not websocket, it is a HTTP Request
	reqHandler, ok := obj.Interface().(RequestHandler)
	if !ok {
		// It should, it cannot be here
		return errors.New("The interface does not implement RequestHandler: " + hc.tip.Name())
	}
	return reqHandler.Handle(c)
}

func (hc *handlerContext) getMiddleRequestHandler(injector *injector) middleRequestHandler {
	// Create a new handler here
	fn := func(c *Context) error {
		obj := reflect.New(hc.tip)
		objElem := obj.Elem()

		err := hc.parseFields(c, objElem, injector)
		if err != nil {
			return err
		}

		if hc.websocket {
			return hc.executeWebsocketHandler(obj, c)
		}
		return hc.executeRequestHandler(obj, c)

	}
	return fn
}
