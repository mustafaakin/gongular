package main

import (
	"github.com/mustafaakin/gongular"
	"time"
	"log"
)

func main() {
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

	err := g.ListenAndServe(":8000")
	if err != nil {
		log.Fatal(err)
	}

	/*
	➜ curl localhost:8000/
	{
	  "Message": "Hello, you are coming from: 127.0.0.1:39018",
	  "Date": "2016-07-09T18:34:23.88065349+03:00"
	}
	*/
}
