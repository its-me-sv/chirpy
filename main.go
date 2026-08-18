package main

import (
	"log"
	"net/http"
)

const (
	port         = "8080"
	filePathRoot = "."
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(filePathRoot)))

	server := &http.Server{
		Handler: mux,
		Addr:    ":" + port,
	}

	log.Printf("Listening to port: %s\n", port)
	log.Fatalln(server.ListenAndServe())
}
