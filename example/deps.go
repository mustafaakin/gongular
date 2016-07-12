package main

import (
	"github.com/mustafaakin/gongular"
	"fmt"
	"math/rand"
)

type User struct {
	Age int
}

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
	g.ProvideCustom(&User{}, func(c *gongular.Context) (error, interface{}){
		u := User{
			Age: rand.Intn(70),
		}

		return nil, &u
	})

	g.GET("/user", func(u *User) string{
		return fmt.Sprintf("Hey, you are %d years old", u.Age)
	})

	g.GET("/db", func(d *DB) string {
		return db.Query("SELECT * FROM mytable")
	})

	g.ListenAndServe(":8000")
}
