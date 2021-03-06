package main

import (
	"log"
	"strconv"
	"net"
)

func carbonListener(address string, port int, done chan bool, metrics []chan *Metric) {
	service := net.JoinHostPort(address, strconv.Itoa(port))
	log.Printf("Carbon listening on %s", service)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error on Accept(): %s", err.Error())
			continue
		}
		go handleMetric(conn, metrics)
	}
	done <- true
}
