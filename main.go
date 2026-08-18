package main

import (
	"log"
	"net/http"
)

const port = "8080"

func main() {
	mux := http.NewServeMux()
	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("Listening to port: %s\n", port)
	log.Fatalln(server.ListenAndServe())
}
