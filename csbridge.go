package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type carbonMetric struct {
	path      string
	value     float64
	timestamp time.Time
}

// Format a carbonMetric as:
//   "pretty": stringified for debug output
//   "statsd": statsd gauge format
//   "carbon": carbon format
func (metric *carbonMetric) Format(format string) string {
	switch format {
	case "statsd":
		return fmt.Sprintf("%s:%f|g", metric.path, metric.value)
	case "carbon":
		return fmt.Sprintf("%s %f %d", metric.path, metric.value, metric.timestamp.Unix())
	default:
		return fmt.Sprintf("path: %s value: %f timestamp: %s", metric.path, metric.value, metric.timestamp.Format(time.RFC3339))
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

// Split a single metric passed in via the  Carbon plain text protocol
// into it's path, value, and timestamp fields
func parseCarbonMetric(msg string) (*carbonMetric, error) {
	metric_split := strings.Split(strings.TrimSpace(msg), " ")
	fmt.Println(metric_split)

	if len(metric_split) < 3 {
		return nil, errors.New("Invalid metric")
	}
	value, _ := strconv.ParseFloat(metric_split[1], 64)
	timestamp_i, _ := strconv.ParseInt(metric_split[2], 10, 64)

	return &carbonMetric{
		path:      metric_split[0],
		value:     value,
		timestamp: time.Unix(timestamp_i, 0)}, nil
}

func handleMetric(conn net.Conn, metrics chan *carbonMetric) {
	conn.SetReadDeadline(time.Now().Add(60 * time.Second)) // set 1 minute timeout
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		log.Printf("Received message: %s\n", msg)

		metric, err := parseCarbonMetric(msg)
		if err != nil {
			log.Printf("%s - Error: %s", strings.Trim(msg, "\n"), err.Error())
			return
		}

		log.Printf("Received metric: %s\n", metric.Format("pretty"))
		metrics <- metric
	}
}

func carbonListener(port int32, done chan bool, metrics chan *carbonMetric) {
	service := fmt.Sprintf(":%d", port)
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	checkError(err)

	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Sprintf("Error on Accept(): %s", err.Error())
			continue
		}
		go handleMetric(conn, metrics)
	}
	done <- true
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("PONG"))
}

func httpListener(port int32, done chan bool) {
	service := fmt.Sprintf(":%d", port)
	http.HandleFunc("/_ping", pingHandler)

	err := http.ListenAndServe(service, nil)
	checkError(err)
	done <- true
}

func statsdWriter(service string, done chan bool, metrics chan *carbonMetric) {
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

func main() {
	done := make(chan bool, 1)
	metrics := make(chan *carbonMetric)
	go carbonListener(2003, done, metrics)
	go statsdWriter("127.0.0.1:8125", done, metrics)

	// for status
	go httpListener(9080, done)
	<-done
}
