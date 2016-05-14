package main

import (
	"fmt"
	"log"
	"net"
)

func statsdWriter(address string, port int, done chan bool, metrics chan *carbonMetric) {
	service := fmt.Sprintf("%s:%d", address, port)

	log.Printf("Sending to statsd at %s", service)
	udpAddr, err := net.ResolveUDPAddr("udp4", service)
	if err != nil {
		panic("Unable to resolve UDP address")
	}

	for {
		metric := <-metrics
		log.Printf("Sending metric: %s", metric.Format("statsd"))
		conn, err := net.DialUDP("udp4", nil, udpAddr)
		if err != nil {
			log.Printf("Error connecting to %s: %s", service, err.Error())
			continue
		}

		defer conn.Close()
		msg := fmt.Sprintf("%s\n", metric.Format("statsd"))
		conn.Write([]byte(msg))
	}
	done <- true
}
