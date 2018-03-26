package main

import (
	"log"
	"net/http"
)

func main() {
	// Simple static webserver:
	log.Fatal(http.ListenAndServe(":7777", http.FileServer(http.Dir("./"))))
}
