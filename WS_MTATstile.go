package main

import (
	"io"
	"net/http"
)

func hello(w http.ResponseWriter, r *http.Request) {
	//w.Header().Set("title","Helloo World")
	io.WriteString(w, "<a href='/link.g?name=blah'>Hello world!</a>")
}

func world(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Name is equal to " + r.URL.Query().Get("name"))
}

func main() {
	http.HandleFunc("/", hello)
	http.HandleFunc("/link.g",world)
	http.ListenAndServe(":8001", nil)
}
