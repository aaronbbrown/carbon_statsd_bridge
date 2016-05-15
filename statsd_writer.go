package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
)

func statsdWriter(address string, port int, done chan bool, metrics chan *Metric, transforms []transform) {
	service := net.JoinHostPort(address, strconv.Itoa(port))

	log.Printf("Sending to statsd at %s", service)
	udpAddr, err := net.ResolveUDPAddr("udp4", service)
	if err != nil {
		panic("Unable to resolve UDP address")
	}

	for {
		metric := <-metrics
		emitMetrics := metric.ApplyTransforms(transforms)

		for _, m := range emitMetrics {
			log.Printf("Sending metrics to statsd: %s", m.Format("statsd"))
			conn, err := net.DialUDP("udp4", nil, udpAddr)
			if err != nil {
				log.Printf("Error connecting to %s: %s", service, err.Error())
				continue
			}

			defer closeConnection(conn)
			msg := fmt.Sprintf("%s\n", m.Format("statsd"))
			conn.Write([]byte(msg))
			closeConnection(conn)
		}
	}
	done <- true
}
