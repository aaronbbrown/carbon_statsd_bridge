package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
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
