![gongular](https://raw.githubusercontent.com/mustafaakin/gongular/master/logo.png)

[![Go Report Card](https://goreportcard.com/badge/github.com/mustafaakin/gongular)](https://goreportcard.com/report/github.com/mustafaakin/gongular)
[![GoDoc](https://godoc.org/github.com/mustafaakin/gongular?status.svg)](https://godoc.org/github.com/mustafaakin/gongular)

**Note:** gongular recently updated, and if you are looking for the previous version it is tagged as [v.1.0](https://github.com/mustafaakin/gongular/tree/v1.0) 

gongular is an HTTP Server Framework for developing APIs easily. It is like Gin Gonic, but it features Angular-like (or Spring like) dependency injection and better input handling. Most of the time, user input must be transformed into a structured data then it must be validated. It takes too much time and is a repetitive work, gongular aims to reduce that complexity by providing request-input mapping with tag based validation.

**Note:** gongular is an opinionated framework and it heavily relies on reflection to achieve these functionality. While there are tests to ensure it works flawlessly, I am open to contributions and opinions on how to make it better. 

## Features

* Automatic Query, POST Body, URL Param binding to structs with easy validation
* Easy and simple dependency injection i.e passing DB connections and other values
* Custom dependency injection with user specified logic, i.e as User struct from a session
* Route grouping that allows reducing duplicated code
* Middlewares that can do preliminary work before routes, groups which might be helpful for authentication checks, logging etc.
* Static file serving 
* Very fast thanks to httprouter

## Simple Usage

gongular aims to be simple as much as possible while providing flexibility. The below example is enough to reply user with its IP.

```go
type WelcomeMessage struct {}
func(w *WelcomeMessage) Handle(c *gongular.Context) error {
    c.SetBody(c.Request().RemoteAddr)
}

g := gongular.NewEngine()
g.GET("/", &WelcomeMessage{})
g.ListenAndServe(":8000")
```

## How to Use

All HTTP handlers in gongular are structs with `Handle(c *gongular.Context) error` function or in other words `RequestHandler` interface, implemented. Request handler objects are flexible. They can have various fields, where some of the fields with specific names are special. For instance, if you want to bind the path parameters, your handler object must have field named `Param` which is a flat struct. Also you can have a `Query` field which also maps to query parameters. `Body` field lets you map to JSON body, and `Form` field lets you bind into form submissions with files.

```go
type MyHandler struct {
    Param struct {
        UserID int       
    }
    Query struct {
        Name  string
        Age   int
        Level float64
    }
    Body struct {
        Comment string
        Choices []string
        Address struct {
            City    string
            Country string
            Hello   string            
        }
    }
}
func(m *MyHandler) Handle(c *gongular.Context) error {
    c.SetBody("Wow so much params")
    return nil
}
```

## Path Parameters

We use julienschmidt/httprouter to multiplex requests and do parametric binding to requests. So the format :VariableName, *somepath is supported in paths. Note that, you can use valid struct tag to validate parameters.

```go
type PathParamHandler struct {
    Param struct {
        Username string
    }
}
func(p *PathParamHandler) Handle(c *Context) error {
    c.SetBody(p.Param.Username)
    return nil
}
```

## Query Parameters

Query parameter is very similar to path parameters, the only difference the field name should be `Query` and it should also be a flat struct with no inner parameters or arrays. Query params are case sensitive and use the exact name of the struct property by default. You can use the `q` struct tag to specify the parameter key

```go
type QueryParamHandler struct {
    Query struct {
        Username string `q:"username"`
        Age int
    }
}
func(p *QueryParamHandler) Handle(c *Context) error {
    println(p.Param.Age)
    c.SetBody(p.Param.Username)
    return nil
}
```

## JSON Request Body 

JSON request bodies can be parsed similar to query parameters, but JSON body can be an arbitrary struct.

```go
type BodyParamHandler struct {
    Body struct {
        Username string
        Age int
        Preferences []string
        Comments []struct {
        	OwnerID int
        	Message string
        }
    }
}
func(p *BodyParamHandler) Handle(c *Context) error {
    println(p.Body.Age)
    c.SetBody(p.Body.Preferences + len(c.Body.Comments))
    return nil
}
```

## Forms and File Uploading

Please note that `Body` and `Form` cannot be both present in the same handler, since the gongular would confuse what to do with the request body.

```go
type formHandler struct {
	Form struct {
		Age      int
		Name     string
		Favorite string
		Fraction float64
	}
}

func (q *formHandler) Handle(c *Context) error {
	c.SetBody(fmt.Sprintf("%d:%s:%s:%.2f",
		q.Form.Age, q.Form.Name, q.Form.Favorite, q.Form.Fraction))
	return nil
}

e.GetRouter().POST("/submit", &formHandler{})
```

### File Uploading

For uploaded files, we use a special struct to hold them in the Form value of the request struct. `UploadedFile` holds the `multipart.File` and the `multipart.Header`, you can do anything you want with them.

```go
type UploadedFile struct {
	File   multipart.File
	Header *multipart.FileHeader
}
```

You can use it in the handler like the following:

```go
type formUploadTest struct {
	Form struct {
		SomeFile     *UploadedFile
		RegularValue int
	}
}

func (f *formUploadTest) Handle(c *Context) error {
	s := sha256.New()
	io.Copy(s, f.Form.SomeFile.File)
	resp := fmt.Sprintf("%x:%d", s.Sum(nil), f.Form.RegularValue)
	c.SetBody(resp)
	return nil
}

e.GetRouter().POST("/upload", &formUploadTest{})
```

## Routes and Grouping

Routes can have multiple handlers, called middleware, which might be useful in grouping the requests and doing preliminary work before some routes. For example, the following grouping and routing is valid:  

```go
type simpleHandler struct{}

func (s *simpleHandler) Handle(c *Context) error {
	c.SetBody("hi")
	return nil
}

// The middle ware that will fail if you supply 5 as a user ID
type middlewareFailIfUserId5 struct {
	Param struct {
		UserID int
	}
}

func (m *middlewareFailIfUserId5) Handle(c *Context) error {
	if m.Param.UserID == 5 {
		c.Status(http.StatusTeapot)
		c.SetBody("Sorry")
		c.StopChain()
	}
	return nil
}

r := e.GetRouter()

g := r.Group("/api/user/:UserID", &middlewareFailIfUserId5{})
g.GET("/name", &simpleHandler{})
g.GET("/wow", &simpleHandler{})

/* 
 The example responses:

 /api/user/5/name -> Sorry 
 /api/user/4/name -> hi
 /api/user/1/wow  -> hi
*/
```


## Field Validation

We use asaskevich/govalidator as a validation framework. If the supplied input does not pass the validation step, http.StatusBadRequest (400) is returned the user with the cause. Validation can be used in Query, Param, Body or Form type inputs. An example can be seen as follows:

```go
type QueryParamHandler struct {
    Query struct {
        Username string `valid:"alpha"`
        Age int
    }
}
func(p *QueryParamHandler) Handle(c *Context) error {
    println(p.Param.Age)
    c.SetBody(p.Param.Username)
    return nil
}
```

If a request with a non valid username field is set, it returns a `ParseError`.

## Dependency Injection

One of the thing that makes gongular from other frameworks is that it provides safe value injection to route handlers. It can be used to store database connections, or some other external utility that you want that to be avilable in your handler, but do not want to make it global, or just get it from some other global function that might pollute the space. Supplied dependencies are provided as-is to route handlers and they are private to supplied router, nothing is global.

### Basic Injection

Gongular allows very basic injection: You provide a value to gongular.Engine, and it provides you to your handler if you want it in your handler function. It is not like a Guice or Spring like injection, it does not resolve dependencies of the injections, it just provides the value, so that you do not use global values, and it makes the testing easier, since you can just test your handler function by mocking the interfaces you like.

```go
type myHandler struct {
	Param struct {
		UserID uint
	}
	Database *sql.DB
}

func (i *myHandler) Handle(c *Context) error {
	c.SetBody(fmt.Sprintf("%p:%d", i.Database, i.Param.UserID))
	return nil
}

db := new(sql.DB)
e.Provide(db)
e.GetRouter().GET("/my/db/interaction/:UserID", &myHandler{})
```

### Keyed Injection

The basic injection works great, but if you want to supply same type of value more than once, you have to use keyed injection so that gongular can differ.

```go
type injectKey struct {
	Val1 int `inject:"val1"`
	Val2 int `inject:"val2"`
}

func (i *injectKey) Handle(c *Context) error {
	c.SetBody(i.Val1 * i.Val2)
	return nil
}

e.ProvideWithKey("val1", 71)
e.ProvideWithKey("val2", 97)

e.GetRouter().GET("/", &injectKey{})
```

### Custom Injection

Sometimes, providing values as is might not be sufficient for you. You can chose to ping the database, create a transaction, get a value from a pool, and these requires implementing a custom logic. Gongular allows you to write a `CustomProvideFunction` which allows you to provide your preferred value with any logic you like.  

```go
type injectCustom struct {
	DB *sql.DB
}

func (i *injectCustom) Handle(c *Context) error {
	c.SetBody(fmt.Sprintf("%p", i.DB))
	return nil
}

e := newEngineTest()

var d *sql.DB
e.CustomProvide(&sql.DB{}, func(c *Context) (interface{}, error) {
    d = new(sql.DB)
    return d, nil
})

e.GetRouter().GET("/", &injectCustom{})
```

### Unsafe Injection

The default `Provide` functions allow you to inject implementations only. Injection of interfaces will not work. During injection, the injector will search for a provided type and fail. For example the following code will not work:

```go
type injectKey struct {
	DB MySQLInterface `inject:"db"`
}

func (i *injectKey) Handle(c *Context) error {
	c.SetBody("yay")
	return nil
}

e.ProvideWithKey("db", &sql.DB{})

e.GetRouter().GET("/", &injectKey{})
```

This will cause an injector error. If you want to inject interfaces you must use `ProvideUnsafe`. `ProvideUnsafe` is a strict key/value injection. You cannot provide multiple values for the same key.

Example usage:


```go
type injectKey struct {
	DB MySQLInterface `inject:"db"`
}

func (i *injectKey) Handle(c *Context) error {
	c.SetBody("yay")
	return nil
}

e.ProvideUnsafe("db", initializeDB())

// This would cause a panic
// e.ProvideUnsafe("db", &sql.DB{})

e.GetRouter().GET("/", &injectKey{})
```


## gongular.Context struct

* `context.SetBody(interface{})` : Sets the response body to be serialized.  
* `context.Status(int)` : Sets the status of the response if not previously set
* `context.MustStatus(int)` : Overrides the previously written status
* `context.Request()` : Returns the underlying raw HTTP Request
* `context.Header(string,string)` : Sets a given response header.   
* `context.Finalize()` : Used to write the response to client, normally should not be used other than in PanicHandler since gongular takes care of the response.
* `context.Logger()` : Returns the logger of the context.

## Route Callback

The route callback, set globally for the engine, allows you to get the stats for the completed request. It contains common info, including the request logs and the matched handlers, how much time it took in each handler, the total time, the total response size written and the final status code, which can be useful for you to send it to another monitoring service, or just some Elasticsearch for log analysis.

```go
type RouteStat struct {
    Request       *http.Request
    Handlers      []HandlerStat
    MatchedPath   string
    TotalDuration time.Duration
    ResponseSize  int
    ResponseCode  int
    Logs          *bytes.Buffer
}
```

## Error Handler

In case you return an error from your function, or another error occurs which makes the request unsatisfiable, `gongular.Engine` calls the error handler function, in which defaults to the following handler:

```go
var defaultErrorHandler = func(err error, c *Context) {
	c.logger.Println("An error has occurred:", err)

	switch err := err.(type) {
	case InjectionError:
		c.MustStatus(http.StatusInternalServerError)
		c.logger.Println("Could not inject the requested field", err)
	case ValidationError:
		c.MustStatus(http.StatusBadRequest)
		c.SetBody(map[string]interface{}{"ValidationError": err})
	case ParseError:
		c.MustStatus(http.StatusBadRequest)
		c.SetBody(map[string]interface{}{"ParseError": err})
	default:
		c.SetBody(err.Error())
		c.MustStatus(http.StatusInternalServerError)
	}

	c.StopChain()
}
```

## WebSockets

Gongular supports websocket connections as well. The handler function is similar to regular route handler interface, but it also allows connection termination if you wish with the `Before` handler.

```go
type WebsocketHandler interface {
	Before(c *Context) (http.Header, error)
	Handle(conn *websocket.Conn)
}
```

First of all, handle function does not return an error, since it is a continuous execution. User is responsible for all the websocket interaction. Secondly, Before is a filter applied just before upgrading the request to websocket. It can be useful for filtering the request and returning an error would not open a websocket but close it with an error. The http.Header is for answering with a http.Header which allows setting a cookie. Can be omitted if not desired.

The nice thing about WebsocketHandler is that it supports Param and Query requests as well, so that all the binding and validation can be done before the request, and you can use it in your handler. 

```go
type wsTest struct {
	Param struct {
		UserID int
	}
	Query struct {
		Track    bool
		Username string
	}
}

func (w *wsTest) Before(c *Context) (http.Header, error) {
	return nil, nil
}

func (w *wsTest) Handle(conn *websocket.Conn) {
	_, msg, err := conn.ReadMessage()
	if err != nil {
		conn.Close()
	}

	toSend := fmt.Sprintf("%s:%d:%s:%t", msg, w.Param.UserID, w.Query.Username, w.Query.Track)
	conn.WriteMessage(websocket.TextMessage, []byte(toSend))
	conn.Close()
}
```
