package main

import (
	api "github.com/Lonor/OpsBot/api"
	"log"
	"net/http"
)

// main For self-hosted HTTP server endpoint
func main() {
	http.HandleFunc("/api/index", api.Handler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
