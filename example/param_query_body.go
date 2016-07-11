package main

import (
	"fmt"
	"github.com/mustafaakin/gongular"
)

type ExampleQuery struct {
	Username string
	Age      int
}

func QueryRequest(q ExampleQuery) string {
	return fmt.Sprintf("Hello, %s, you are %d years old.", q.Username, q.Age)
}

type ExampleParam struct {
	Username string
}

func ParamRequest(p ExampleParam) string {
	return "Hi: " + p.Username
}

type ExampleBody struct {
	Username string
	Choices  []struct {
		Question string
		Answer   string
	}
}

func BodyRequest(b ExampleBody) string {
	s := "Hi , " + b.Username + ". "
	for _, choice := range b.Choices {
		s += choice.Question + ": " + choice.Answer + "; "
	}
	return s
}

func main() {
	r := gongular.NewRouter()
	r.GET("/query", QueryRequest)
	/*
	   ➜ curl "localhost:8000/query?Username=Mustafa&Age=25"
	   "Hello, Mustafa, you are 25 years old."
	*/
	r.GET("/param/:Username", ParamRequest)
	/*
	   ➜ curl localhost:8000/param/Mustafa
	   "Hi: Mustafa"
	*/
	r.POST("/body", BodyRequest)
    /*
    ➜  curl -H "Content-Type: application/json" -X POST -d '{"Username":"Mustafa","Choices": [{"Question":"How old are you?","Answer":"25"}, {"Question":"What is your favorite color?", "Answer":"Blue"}]}' http://localhost:8000/body
    "Hi , Mustafa. How old are you?: 25; What is your favorite color?: Blue; "
    */
	r.ListenAndServe(":8000")
}
