package main

import (
	"fmt"
	"log"
	"net/http"
)

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("PONG"))
}

func httpListener(address string, port int, done chan bool) {
	service := fmt.Sprintf("%s:%d", address, port)

	http.HandleFunc("/_ping", pingHandler)
	err := http.ListenAndServe(service, nil)
	checkError(err)
	log.Printf("HTTP listening on %s", service)

	done <- true
}
