![gongular](https://raw.githubusercontent.com/mustafaakin/gongular/master/logo.png)

[![Coverage Status](https://coveralls.io/repos/github/mustafaakin/gongular/badge.svg?branch=master)](https://coveralls.io/github/mustafaakin/gongular?branch=master)
[![Build Status](https://travis-ci.org/mustafaakin/gongular.svg?branch=master)](https://travis-ci.org/mustafaakin/gongular)
[![Go Report Card](https://goreportcard.com/badge/github.com/mustafaakin/gongular)](https://goreportcard.com/report/github.com/mustafaakin/gongular)
[![GoDoc](https://godoc.org/github.com/mustafaakin/gongular?status.svg)](https://godoc.org/github.com/mustafaakin/gongular)

gongular is an HTTP Server Framework for developing APIs easily. It is like Gin Gonic, but it features Angular-like (or Spring like) dependency injection and better input handling. Most of the time, user input must be transformed into a structured data then it must be validated. It takes too much time and is a repetitive work, gongular aims to reduce that complexity by providing request-input mapping with tag based validation.

## Features

* Automatic Query, POST Body, URL Param binding to structs with easy validation
* Easy and simple dependency injection i.e passing DB connections and other values
* Custom dependency injection with user specified logic, i.e as User struct from a session
* Route grouping that allows reducing duplicated code
* Middlewares that can do preliminary work before routes, groups which might be helpful for authentication checks, logging etc.
* Static file serving 
* Very fast thanks to httprouter

## Simple Usage

gongular aims to be simple as much as possible while providing flexibility. The below example is enough to reply user with its IP.

```go
type WelcomeMessage struct {
  Message string
  Date    time.Time
}

g := gongular.NewRouter()
g.GET("/", func(c *gongular.Context) WelcomeMessage {
  return WelcomeMessage{
    Message: "Hello, you are coming from: " + c.Request().RemoteAddr,
    Date:    time.Now(),
  }
})
```

And output is:

```
[INFO ] 2016/07/09 18:34:16 GET   /                                        func(*gongular.Context) main.WelcomeMessage
[INFO ] 2016/07/09 18:34:16 Listening HTTP on :8000
```

When you make a request, you will see how much time passed in your handler, and bytes served and total time including JSON encoding.

```zsh
➜ curl localhost:8000/
{
  "Message": "Hello, you are coming from: 127.0.0.1:39018",
  "Date": "2016-07-09T18:34:23.88065349+03:00"
}
```

```zsh
[DEBUG] 2016/07/09 18:34:23 GET   /                              <func(*gongular.Context) main.WelcomeMessage Value>   33.004µs
[INFO ] 2016/07/09 18:34:23 GET   /                              /                                                     93.106µs  200 110
```

## How to install 

You can just go get it via `go get github.com/mustafaakin/gongular`. Requires Go >= 1.6

## Handler Function Format

Route handler functions are flexible. They can have various parameters or output values. While registering the handler functions, they are examined via reflection, and saved properly to bind the requests or answer them with appropriate output. This feature might not work for everyone, but it is what makes gongular different from other frameworks.

Allowed **input types**:

* `*gongular.Context`   : Wrapper for http request and http response writer, that has some useful utilities
* `SomethingBody`       : if given struct's name ends with body, it binds HTTP request body, treating it as a JSON
* `SomethingQuery`      : if given struct's name ends with query, it bind the query params 
* `SomethingParam`      : if given struct's name ends with param, it binds the URL params

Allowed **output types**:

* `error`  : An internal error that displays an error message to request and logs details in console
* `any`    : Renders the given value as a JSON to user

So, the following is completely valid, you can use any number of inputs or outputs. 

```go
func (c *gongular.Context, body SomethingBody, query SomethingQuery, param SomethingParam) (error, SomethingOutput){
    
}
```

The following is also valid, which does not take or return anything: 

```go
func() {

}
```

You can use the combinations as you might need:

```go
func(c *gongular.Context, s SometingQuery, p SomethingParam) SomeResponse {

}
```

## Routes, Middlewares and Grouping

Routes can have multiple handlers, called middlewares, which might be useful in grouping the requests and doing preliminary work before some routes. For example, the following grouping and routing is valid:

```go
func Logger(c *gongular.Context) {
	log.Println("Printed before every request")
}

func Index() string {
	return "Hello, world"
}

func SomePath() int {
	return 42
}

r := gongular.NewRouter()
g1 := r.Group("/", Logger)
{
    g1.GET("/", Index)
    g1.GET("/answer", SomePath)
    g2 := g1.Group("/admin")
    {
        g2.GET("/delete", Logger, Index)
    }
}

r.ListenAndServe(":8000")
```

Output will be:

```zsh
➜ curl localhost:8000  # Logger, Index Called
"Hello, world"
➜ curl localhost:8000/answer # Logger, SomePath called
42
➜ curl localhost:8000/admin/delete   # Logger, Logger, Index called
"Hello, world"
```

## Query Parameters

```go
type MyQuery struct {
	Username string
	Age      int
}

func QueryRequest(q MyQuery) string{
	return fmt.Sprintf("Hi: %s, You are %d years old", q.Username, q.Age)
}

func main() {
	r := gongular.NewRouter()
	r.GET("/query", QueryRequest)
	r.ListenAndServe(":8000")
}
```

And the output will be:

```zsh
➜ curl "localhost:8000/query?Username=Mustafa&Age=25"
"Hi: Mustafa, You are 25 years old"
```

## Path Parameters

We use [julienschmidt/httprouter](https://github.com/julienschmidt/httprouter) to multiplex requests and do parametric binding to requests. So the format `:VariableName, *somepath` is supported in paths. Note that, you can use `valid` struct tag to validate parameters.

Also, note that, the struct name must end with **Param**

```go
type MyParam struct {
	Username string
}

func AnotherRequest(u MyParam) string{
	return "Hi: " + u.Username
}

func main() {
	r := gongular.NewRouter()
    	r.GET("/param/:Username", AnotherRequest)
	r.ListenAndServe(":8000")
}
```

And the output will be:

```zsh
➜ curl localhost:8000/param/Mustafa
"Hi: Mustafa"
```

## POST JSON Body

```go
type MyLongBody struct {
	Username string
	Choices []struct {
		Question string
		Answer   string
	}
}

func ParseThatBodyPlease(b MyLongBody) string {
	str := "Hi " + b.Username + ". "
	for _, choice := b.Choices {
		str + = choice.Question + ":" + choice.Answer + "; " 
	}
	return str
}

func main() {
	r := gongular.NewRouter()
	r.POST("/body", ParseThatBodyPlease)
	r.ListenAndServe(":8000")
}
```

And the output will be:

```zsh
➜ curl -H "Content-Type: application/json" -X POST -d '{"Username":"Mustafa","Choices": [{"Question":"How old are you?","Answer":"25"}, {"Question":"What is your favorite color?", "Answer":"Blue"}]}' http://localhost:8000/body
"Hi , Mustafa. How old are you?: 25; What is your favorite color?: Blue; "
```

## Validation

We use [asaskevich/govalidator](https://github.com/asaskevich/govalidator) as a validation framework. If the supplied input does not pass the validation step, `http.StatusBadRequest (400)` is returned the user with the cause. Validation can be used in *Query*, *Param* or *Body* type inputs.

```go
type RegisterBody struct {
	Username string	`valid:"alphanum,required"`
	Password string `valid:"required"`
}

func main(){
	r := gongular.NewRouter()
	r.POST("/register", func(b RegisterBody) string {
		return "Saved succesfully"
	})
	r.ListenAndServe(":8000")
}
```

Invalid request, password not supplied #1:

```zsh
➜ curl -H "Content-Type: application/json" -X POST -d '{"Username":"Mustafa"}' http://localhost:8000/register
"Submitted body is not valid: Password: non zero value required;"
```

Invalid request, non alpha numeric username:

```zsh
➜ curl -H "Content-Type: application/json" -X POST -d '{"Username":"!!!Mustafa","Password": "123"}' http://localhost:8000/register
"Submitted body is not valid: Username: !!!Mustafa does not validate as alphanum;"%
```

Valid request:

```zsh
➜ curl -H "Content-Type: application/json" -X POST -d '{"Username":"Mustafa","Password": "123"}' http://localhost:8000/register
"Saved succesfully"
```

## Dependencies

One of the thing that makes `gongular` from other frameworks is that it provides safe value injection to route handlers. It can be used to store database connections, or some other external utility that you want that to be avilable in your handler, but do not want to make it global, or just get it from some other global function that might pollute the space. Supplied dependencies are provided as-is to route handlers and they are private to supplied router, nothing is global.  

```go
type DB struct {
	Hostname string
	Database string
}

func (d *DB) Query(sql string) string {
	return fmt.Sprintf("Host=%s DB=%s SQL=%s", d.Hostname, d.Database,sql)
}

func main(){
	g := gongular.NewRouter()
	db := &DB{
		Hostname: "server0",
		Database: "testdb",
	}
	g.Provide(db)
	g.GET("/db", func(d2 *DB) string {
		return d2.Query("SELECT * FROM mytable")
	})
	g.ListenAndServe(":8000")
}
```

And the output will be:

```zsh
➜ curl localhost:8000/db
"Host=server0 DB=testdb SQL=SELECT * FROM mytable"
```

## Custom Dependencies

Static dependencies are great, but they do not provide more flexibility. They are static, and supplied as-is. If you want custom logic while providing your dependency, i.e. providing username from session, you can use the following:

```go
type User struct {
	Age int
}

func main(){
	g := gongular.NewRouter()
	g.ProvideCustom(&User{}, func(c *gongular.Context) (error, interface{}){
		u := User{
			Age: rand.Intn(70),
		}
		return nil, &u
	})
	g.GET("/user", func(u *User) string{
		return fmt.Sprintf("Hey, you are %d years old", u.Age)
	})
	g.ListenAndServe(":8000")
}
```

And the output will be:

```zsh
➜ curl localhost:8000/user
"Hey, you are 41 years old"
➜ curl localhost:8000/user
"Hey, you are 37 years old"
```

Note that, errors are used for indicating internal errors. If you supply a value to error, the gongular router will write 500 as a status. If you want to indicate that you could not supply a value, you have to proivde nil as second output.

## `gongular.Context` struct

gongular.Context is a wrapper for http.Request and http.ResponseWriter and contains useful utilities. 

* `context.Status(int)`: Sets the status of a response if not set previously
* `context.StopChain()`: Used to stop the next handlers to be executed, useful in middlewares
* `context.Header(key,value)`: Sets the HTTP Header (key) to value
* `context.Finalize()`: Used to write the response to client, normally should not be used other than in PanicHandler since gongular takes care of the response.

## Logging

Debugging can be toggled with `router.DisableDebug()` and `router.EnableDebug()`. If desired, `INFO` and `DEBUG` loggers of a router can be changed as follows: 

```go
router.DebugLog = log.New(/* valid options */) /* or custom loggers that implements log.Logger interface like logrus */
router.InfoLog = log.New(/* valid options */)
```

## Custom Error and Panic Handling

If your HandlerFunction returns a non-nil error, the other response is discarded and an error handler is executed (same for panics). Default errors and panics handles writes the causes to response. However, this might not be the thing you want all the time, some errors are better not to be shown to users. To change default error handlers, you can use the following:

```go
r := gongular.NewRouter()
r.ErrorHandler = func(err error, c *gongular.Context){  
    c.StopChain()
    c.MustStatus(http.StatusInternalServerError)
    c.SetBody(map[string]string{
        "error": err.Error(),
    })
}

r.PanicHandler = func(v interface{}, c *Context) {
    c.StopChain()
    c.MustStatus(http.StatusInternalServerError)
    c.SetBody(map[string]interface{}{
        "panic": v,
    })
    // Note that you have to call context.Finalize method to write 
    // the response since panic causes interuption to regular 
    // gongular flow.
    c.Finalize()
}

r.GET("/", Index)
```


## TODO

gongular is relatively new, and following needs to be completed:

* ~~Unit tests~~
* ~~Integration tests~~
* Benchmarks

## Contribute

You are welcome to contribute in anyway possible. However, I would appreciate a discussion before just sending PRs to agree on something.
