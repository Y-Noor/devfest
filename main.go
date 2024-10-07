package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

func main() {
	fmt.Println("hello")

	h1 := func(w http.ResponseWriter, r *http.Request) {
		templ := template.Must(template.ParseFiles("index.html"))
		templ.Execute(w, nil)
	}

	http.HandleFunc("/", h1)
	log.Fatal(http.ListenAndServe(":8000", nil))

}
