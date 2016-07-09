package main

import (
	"errors"
	"log"
	"time"

	"github.com/mustafaakin/gongular"
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

type TestQuery struct {
	Username string
	Age      int
}

type LoginBody struct {
	Username string
	Password string
}

type RegisterBody struct {
	Username string `valid:"alphanum"`
	Password string `valid:"numeric"`
}

type RegisterResponse struct {
	Key  string
	Time time.Time
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

	r.GET("/canYouDrive", func(q TestQuery) string {
		if q.Age < 18 {
			return q.Username + ", you are young, sorry. No wheels."
		} else {
			return "Hey " + q.Username + ", you are a grown up, do what you want."
		}
	})

	r.POST("/login", func(b LoginBody) bool {
		return b.Username == "mustafa" && b.Password == "123"
	})

	r.POST("/register", func(b RegisterBody) RegisterResponse {
		return RegisterResponse{
			Key:  b.Username + "-" + b.Password,
			Time: time.Now(),
		}
	})

	r.GET("/error", func() error {
		return errors.New("It' s a trap")
	})

	a := r.Group("/admin", func(c *gongular.Context) {
		log.Println("Dangerous action, admin stuff..")
	})
	{
		a.GET("/exterminate", func() string {
			return "Exterminated"
		})
	}

	r.Static("/assets", "example/static")

	// Default listen and serve
	err := r.ListenAndServe(":8000")
	if err != nil {
		log.Fatal(err)
	}
}
