package main

import (
	"log"
	"net"
	"net/http"
	"strconv"
)

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("PONG"))
}

func httpListener(address string, port int, done chan bool) {
	service := net.JoinHostPort(address, strconv.Itoa(port))

	http.HandleFunc("/_ping", pingHandler)
	err := http.ListenAndServe(service, nil)
	checkError(err)
	log.Printf("HTTP listening on %s", service)

	done <- true
}
