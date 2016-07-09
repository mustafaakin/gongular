package main

import (
	"github.com/mustafaakin/gongular"
	"log"
)

type UserSession struct {
	Username string
}

type UserSession2 struct {
	Username string
}

type UserParam struct {
	Username string
}

func main() {
	// Create a new Router, currently no options required
	r := gongular.NewRouter()
	r.DisableDebug()
	r.GET("/", func(c *gongular.Context) string {
		return "Hello" + c.Request().UserAgent()
	})

	r.GET("/user/:Username", func(u UserParam) string {
		return "Hi " + u.Username
	})

	r.Static("/assets", "example/static")

	// Default listen and serve
	err := r.ListenAndServe(":8000")
	if err != nil {
		log.Fatal(err)
	}
}
