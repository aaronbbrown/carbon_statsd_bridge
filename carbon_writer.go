package main

import (
	"fmt"
	"log"
	"net"
	"time"
	"strconv"
)

func connectCarbon(address string, port int) (net.Conn, error) {
	service := net.JoinHostPort(address, strconv.Itoa(port))

	log.Printf("Connecting to carbon at %s", service)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTCP("tcp4", nil, tcpAddr)
	if err != nil {
		return nil, err
	}
	return conn, err
}

func closeConnection(conn net.Conn) {
	if conn != nil {
		conn.Close()
	}
}

func carbonWriter(address string, port int, done chan bool, metrics chan *Metric) {
	for {
		// keep attempting to reconnect to carbon
		conn, err := connectCarbon(address, port)
		if err == nil {
			log.Printf("Connected to carbon %s:%d", address, port)
		} else {
			log.Printf("Error connecting to carbon %s:%d: %s", address, port, err.Error())
		}
		defer closeConnection(conn)

		for conn != nil {
			metric := <-metrics
			msg := fmt.Sprintf("%s\n", metric.Format("carbon"))
			log.Printf("Sending metric to carbon: %s", msg)

			_, err := conn.Write([]byte(msg))
			if err != nil {
				log.Printf("Error sending metric to carbon %s:%d: %s", address, port, err.Error())
				closeConnection(conn)
				break
			}
		}

		// attempt reconnection every second
		time.Sleep(1 * time.Second)
	}
	done <- true
}
