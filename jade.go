package gongular

import (
	"strings"
	"path"
	"io/ioutil"
	"github.com/Joker/jade"
	"fmt"
	"net/http"
)

type p0Param struct {
	P0 string
}

type p1Param struct {
	P0, P1 string
}

type p2Param struct {
	P0, P1, P2 string
}

type p3Param struct {
	P0, P1, P2, P3 string
}

type JadeRenderer struct {
	BasePath string
}

func (j *JadeRenderer) renderSingle(w http.ResponseWriter, template string){
	newPath := path.Join(j.BasePath, template)

	// Read the jade file
	file, err := ioutil.ReadFile(newPath)
	if err != nil {
		fmt.Println(err)
	}
	k := string(file)

	// Parse the template
	tpl, err := jade.Parse(template, k)
	if err != nil {
		fmt.Println(err)
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, tpl)
}

func(j *JadeRenderer) RegisterSingle(g *Router, path, template string){
	g.GET(path, func(w http.ResponseWriter) {
		j.renderSingle(w, template)
	})
}


func (j *JadeRenderer) render(w http.ResponseWriter, arr ...string) {
	a := ""
	for i := 0; i < len(arr)-1; i++ {
		a += arr[i] + "/"
	}

	last := arr[len(arr)-1]
	last = strings.Replace(last, ".html", ".jade", -1)
	a += last

	newPath := path.Join(j.BasePath, a)

	// Read the jade file
	file, err := ioutil.ReadFile(newPath)
	if err != nil {
		fmt.Println(err)
		//		return "", err
	}

	// Convert thed file to string
	k := string(file)

	// Parse the template
	tpl, err := jade.Parse(a, k)
	if err != nil {
		fmt.Println(err)

		//		return "", err
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, tpl)
}

func (j *JadeRenderer) RegisterRoutes(g *Router) {
	g.GET("/partials/{P0}", func(w http.ResponseWriter, p p0Param) {
		j.render(w, "partials", p.P0)
	})
	g.GET("/partials/{P0}/{P1}", func(w http.ResponseWriter, p p1Param) {
		j.render(w, "partials", p.P0, p.P1)
	})
	g.GET("/partials/{P0}/{P1}/{P2}", func(w http.ResponseWriter, p p2Param) {
		j.render(w, "partials",p.P0, p.P1, p.P2)
	})
	g.GET("/partials/{P0}/{P1}/{P3}", func(w http.ResponseWriter, p p3Param) {
		j.render(w, "partials",p.P0, p.P1, p.P2, p.P3)
	})
}
