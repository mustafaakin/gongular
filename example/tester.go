package main

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"github.com/mustafaakin/gongular"
	"log"
	"net/http"
	"os"
	"fmt"
)

type UserSession struct {
	Username string
}

type UserSession2 struct {
	Username string
}

func main() {
	// Not very important, just to see proper colored log output in Intellij IDEA
	logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})
	logrus.SetOutput(os.Stdout)
	log.SetOutput(os.Stdout)

	// Create a new Router, currently no options required
	r := gongular.NewRouter()
	r.GET("/", func(r *http.Request) string {
		return "Hello: " + r.RemoteAddr
	})

	r.ProvideCustom(UserSession{}, func(w http.ResponseWriter, r *http.Request) (error, interface{}) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Sorry but you are unauthorized")
		return nil, nil
	})

	r.ProvideCustom(UserSession2{}, func(w http.ResponseWriter, r *http.Request) (error, interface{}) {
		return errors.New("could not connect to db"), nil
	})

	r.GET("/provideFail", func(u UserSession) string {
		return "Username: " + u.Username
	})

	r.GET("/provideFail2", func(u UserSession2) string {
		return "Username: " + u.Username
	})

	r.Static("example/static")

	// Default listen and serve
	err := r.ListenAndServe(":8000")
	if err != nil {
		logrus.Fatal(err)
	}
}
