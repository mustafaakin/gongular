![gongular](https://raw.githubusercontent.com/mustafaakin/gongular/master/logo.png)

gongular is an HTTP Server Framework for developing API. It is like Gin Gonic, but it features Angular-like (or Spring like) dependency injection and better input handling. Simple usage is as follows

## Simple Usage

gongular aims to be simple as much as possible while providing flexibility. The below example is enough to reply user with its IP.

```go
r := gongular.NewRouter()
r.GET("/", func(r *http.Request) string{
    return "Hello: " + r.RemoteAddr
})

r.ListenAndServe(":8000")
```

And output is:

```
INFO[0000] Handler registerging                          handler=func(*http.Request) string index=0 method=GET path=/
INFO[0000] Listening on HTTP                             address=:8000
```

## How to install 

You can just go get it via `go get github.com/mustafaakin/gongular`

## Handler Function Format

Route handler functions are flexible. They can have various parameters or output values. While registering the handler functions, they are examined via reflection, and saved properly to bind the requests or answer them with appropriate output. This feature might not work for everyone, but it is what makes gongular different from other frameworks.

Allowed **input types**:

* `*http.Request`       : original http request object
* `http.ResponseWriter` : you can manually use the response writer if you wish but should not be needed in most cases
* `SomethingBody`       :  if given struct's name ends with body, it binds HTTP request body, treating it as a JSON
* `SomethingQuery`      : if given struct's name ends with query, it bind the query params 
* `SomethingParam`      : if given struct's name ends with param, it binds the URL params

Allowed **output types**:

* `int`              : Indicates a HTTP status code, if -1, ignored
* `boolean`          : Stops request handler chain, useful in grouping routes
* `error`            : An internal error that displays an error message to request and logs details in console
* `struct or string` : Renders the given struct/string as JSON to user

So, the following is completly valid, you can use any number of inputs or outputs. 

```go
func (w http.ResponseWriter, r *http.Request, body SomethingBody, query SomethingQuery, param SomethingParam) (int, boolean, error, SomethingOutput){
    
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

## Path Parameters

We use [gorilla/mux](https://github.com/gorilla/mux) to multiplex requests and do parametric binding to requests. So the format `{VariableName}` is supported in paths. Note that, you can use `valid` struct tag to validate parameters. We use [asaskevich/govalidator](https://github.com/asaskevich/govalidator) as a validation framework. If the supplied input does not pass the validation step, `http.StatusBadRequest (400)` is returned the user with the cause.

Also, note that, the struct name must end with **Param**

```go
type GetUserParam struct {
    Username string    `valid:"alphanum"`
}
r := gongular.NewRouter()
r.GET("/user/{Username}", func(param GetUserParam) (int){
    if param.Username == "mustafa" {
        return http.StatusOK
    } else {
        return http.StatusUnauthorized
    }
})
```

The above code will match the URL requests `/user/mustafa` and `/user/someonelese` and return status codes accordingly.

## Query Parameters

Query parameters are commonly used in the HTTP Requests. It is so similar to path parameters, however the struct name must end with **Query**

```go
type CanYouDriveQuery struct {
    Username string    `valid:"alphanum"`
    Age      int 
}
r := gongular.NewRouter()
r.GET("/canYouDrink", func(c CanYouDriveQuery) (string){
    if c.Age < 18  {
        return "Sorry " + c.Username, " you cannot drive"
    } else {
        return "Okay, you can drive"
    }
})
```

## POST JSON Body

In POST requests, you can also use the JSON body to map it as a struct. The name of the supplied input parameter must end with **Body** to do so.

```go
type UpdateMyChoicesBody struct {
    Username  string    `valid:"alphanum"`
    Password  string
    Choices[] string
}

type UpdateMyChoicesResponse struct {
    StatusMessage string
    Elapsed       int
}

r := gongular.NewRouter()
r.POST("/update/choices", func(u UpdateMyChoicesBody) (UpdateMyChoicesResponse){
    // some business logic
    return UpdateMyChoicesResponse{
       StatusMessage: "Some of your choices updated.",
       Elapsed: 50,
    }
})
```

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
