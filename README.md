![gongular](https://raw.githubusercontent.com/mustafaakin/gongular/master/logo.png)
[![Go Report Card](https://goreportcard.com/badge/github.com/mustafaakin/gongular)](https://goreportcard.com/report/github.com/mustafaakin/gongular)
[![GoDoc](https://godoc.org/github.com/mustafaakin/gongular?status.svg)](https://godoc.org/github.com/mustafaakin/gongular)

gongular is an HTTP Server Framework for developing APIs easily. It is like Gin Gonic, but it features Angular-like (or Spring like) dependency injection and better input handling. Most of the time, user input must be transformed into a structured data then it mus be validated. It takes too much time and is a repetitive work, gongular aims to reduce that complexity by providing request-input mapping with tag based validation.

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

```
➜ curl localhost:8000/
{
  "Message": "Hello, you are coming from: 127.0.0.1:39018",
  "Date": "2016-07-09T18:34:23.88065349+03:00"
}

[DEBUG] 2016/07/09 18:34:23 GET   /                              <func(*gongular.Context) main.WelcomeMessage Value>   33.004µs
[INFO ] 2016/07/09 18:34:23 GET   /                              /                                                     93.106µs  200 110
```

## How to install 

You can just go get it via `go get github.com/mustafaakin/gongular`

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

So, the following is completly valid, you can use any number of inputs or outputs. 

```go
func (c *gongular.Context, body SomethingBody, query SomethingQuery, param SomethingParam) (error, SomethingOutput){
    
}
```

## Routes, Middlewares and Grouping

Routes can have multiple handlers, called middlewares, which might be useful in grouping the requests and doing preliminary work before some routes. For example, the following grouping and routing is valid:

```go
r := gongular.NewRouter()
g := r.Group("/admin", CheckAdminAuth)
{
    users := g.Group("/users")
    {
        users.GET("/list", ListUsers)
        users.POST("/delete/:user", LogDangeroursAction, DeleteUser)
    }
    g.POST("/stopSystem", MailOthers, KillTheLights,  StopSystem)
}
```

## Query Parameters

```go
func main() {
	r := gongular.NewRouter()
	r.GET("/query", QueryRequest)
	r.ListenAndServe(":8000")
}
```

And the output will be:

```zsh
➜ curl localhost:8000/param/Mustafa
"Hi: Mustafa"
```

## Path Parameters

We use [julienschmidt/httprouter](https://github.com/julienschmidt/httprouter) to multiplex requests and do parametric binding to requests. So the format `:VariableName, *somepath` is supported in paths. Note that, you can use `valid` struct tag to validate parameters. We use [asaskevich/govalidator](https://github.com/asaskevich/govalidator) as a validation framework. If the supplied input does not pass the validation step, `http.StatusBadRequest (400)` is returned the user with the cause.

Also, note that, the struct name must end with **Param**

```go
func main() {
	r := gongular.NewRouter()
    r.GET("/param/:Username", ParamRequest)
	r.ListenAndServe(":8000")
}
```

And the output will be:

```zsh
➜ curl localhost:8000/param/Mustafa
"Hi: Mustafa"
```

## POST JSON Body

```zsh
func main() {
	r := gongular.NewRouter()
	r.POST("/body", BodyRequest)
	r.ListenAndServe(":8000")
}
```

And the output will be:

```zsh
➜ curl -H "Content-Type: application/json" -X POST -d '{"Username":"Mustafa","Choices": [{"Question":"How old are you?","Answer":"25"}, {"Question":"What is your favorite color?", "Answer":"Blue"}]}' http://localhost:8000/body
"Hi , Mustafa. How old are you?: 25; What is your favorite color?: Blue; "
```

## Validation

We use govalidator. Details soon.

## Dependencies

One of the thing that makes `gongular` from other frameworks is that it provides safe value injection to route handlers. It can be used to store database connections, or some other external utility that you want that to be avilable in your handler, but do not want to make it global, or just get it from some other global function that might pollute the space. Supplied dependencies are provided as-is to route handlers and they are private to supplied router, nothing is global.  

```go
type User struct {
    Username   string
    Email      string
    Age        int
    IsActive   bool
}
db = sqlx.MustConnect("sqlite3", ":memory:")
r := gongular.NewRouter()
r.Proivde(db)
r.GET("/users", func(db *sqlx.DB) ([]Users, error){
    var users []Users
    err = db.Select(&pp, "SELECT * FROM users")
    return users, err
})
```

## Custom Dependencies

Static dependencies are great, but they do not provide more flexibility. They are static, and supplied as-is. If you want custom logic while providing your dependency, i.e. providing username from session, you can use the following:

```go
r := gongular.NewRouter()
r.ProvideCustom(UserSession{}, func(w http.ResponseWriter, r *http.Request) (error, interface{}) {
    session, err := store.Get(r, "session-name")
    if err != nil {
        return err, nil
    }
    if val, ok := session.Values["username"];ok {
        return nil, UserSession{
            Username: val,
        }
    } else {
        w.WriteHeader(http.StatusUnauthorized)
        w.Header().Set("Content-Type", "application/json")
        fmt.Fprintf(w, "You are unauthorized!!!")
        return nil, nil
    }
})
```

Note that, errors are used for indicating internal errors. If you supply a value to error, the gongular router will write 500 as a status. If you want to indicate that you could not supply a value, you have to proivde nil as second output.

## Logging

[Sirupsen/logrus](https://github.com/Sirupsen/logrus) is used in logging and will be more configurable in following releases.

## TODO

* Code refactoring for better readability & testability
* Actual tests
* Static file serving
* Configurable logging
* Better validation
* Better info about routes
* Stats about route performance (not really needed)

## Contribute

You are welcome to contribute in anyway possible. However, I would appreciate a discussion before just sending PRs to agree on something.
