package main

import "github.com/mustafaakin/gongular"

type RegisterBody struct {
	Username string	`valid:"alphanum,required"`
	Password string `valid:"required"`
}

func main(){
	r := gongular.NewRouter()
	r.POST("/register", func(b RegisterBody) string {
		return "Saved succesfully"
	})
	r.ListenAndServe(":8000")
	/*
	➜ curl -H "Content-Type: application/json" -X POST -d '{"Username":"Mustafa"}' http://localhost:8000/register
    "Submitted body is not valid: Password: non zero value required;"

    ➜ curl -H "Content-Type: application/json" -X POST -d '{"Username":"!!!Mustafa","Password": "123"}' http://localhost:8000/register
    "Submitted body is not valid: Username: !!!Mustafa does not validate as alphanum;"%

    ➜ curl -H "Content-Type: application/json" -X POST -d '{"Username":"Mustafa","Password": "123"}' http://localhost:8000/register
    "Saved succesfully"
	 */
}
