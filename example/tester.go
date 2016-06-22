package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/mustafaakin/gongular"
	"log"
	"net/http"
	"os"
)

func main() {
	// Not very important, just to see proper colored log output in Intellij IDEA
	logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})
	logrus.SetOutput(os.Stdout)
	log.SetOutput(os.Stdout)

	// Create a new Router, currently no options required
	r := gongular.NewRouter()
	r.GET("/", func(r *http.Request) string{
		return "Hello: " + r.RemoteAddr
	})

	// Default listen and serve
	err := r.ListenAndServe(":8000")
	if err != nil {
		logrus.Fatal(err)
	}
}