package main

import (
	"github.com/mustafaakin/gongular"
	"log"
)

func Logger(c *gongular.Context) {
	log.Println("Printed before every request")
}

func Index() string {
	return "Hello, world"
}

func SomePath() int {
	return 42
}

func main() {
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

	/*
    ➜  ~ curl localhost:8000  # Logger, Index Called
    "Hello, world"
    ➜  ~ curl localhost:8000/answer # Logger, SomePath called
    42
    ➜  ~ curl localhost:8000/admin/delete   # Logger, Logger, Index called
    "Hello, world"
	 */
}
