package main

import (
	"LuG/planets/api"
	"log"
	"net/http"
)

func main() {
	log.Fatal(
		"\"main\" package, web server start-up failed --> ",
		http.ListenAndServe(":8080", api.Router()),
	)
}
