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

type Metric struct {
	path      string
	value     float64
	timestamp time.Time
}

// Format a Metric as:
//   "pretty": stringified for debug output
//   "statsd": statsd gauge format
//   "carbon": carbon format
func (metric *Metric) Format(format string) string {
	switch format {
	case "statsd":
		return fmt.Sprintf("%s:%f|g", metric.path, metric.value)
	case "carbon":
		return fmt.Sprintf("%s %f %d", metric.path, metric.value, metric.timestamp.Unix())
	default:
		return fmt.Sprintf("path: %s value: %f timestamp: %s", metric.path, metric.value, metric.timestamp.Format(time.RFC3339))
	}
}

// transform a metric using transform rules.  Returns a new metric
// or an error if the metric didn't match the transform
func (metric Metric) Transform(t *transform) (Metric, error) {
	if t.Regexp.MatchString(metric.path) {
		metric.path = t.Regexp.ReplaceAllString(metric.path, t.Replace)
		return metric, nil
	} else {
		return metric, errors.New("No match")
	}
}

// iterate over all the transforms and apply them against the metric
// if no transforms match, emit the original, unaltered metric
func (metric *Metric) ApplyTransforms(transforms []transform) []*Metric {
	var emitMetrics []*Metric
	for _, t := range transforms {
		newMetric, err := metric.Transform(&t)
		if err == nil {
			emitMetrics = append(emitMetrics, &newMetric)
		}
	}

	// if no transforms matched, emit the original metric
	if len(emitMetrics) < 1 {
		emitMetrics = append(emitMetrics, metric)
	}

	return emitMetrics
}

// Split a single metric passed in via the  Carbon plain text protocol
// into it's path, value, and timestamp fields
func parseMetric(msg string) (*Metric, error) {
	metric_split := strings.Split(strings.TrimSpace(msg), " ")

	if len(metric_split) < 3 {
		return nil, errors.New("Invalid metric")
	}
	value, _ := strconv.ParseFloat(metric_split[1], 64)
	timestamp_i, _ := strconv.ParseInt(metric_split[2], 10, 64)

	return &Metric{
		path:      metric_split[0],
		value:     value,
		timestamp: time.Unix(timestamp_i, 0)}, nil
}

func handleMetric(conn net.Conn, metrics []chan *Metric) {
	conn.SetReadDeadline(time.Now().Add(60 * time.Second)) // set 1 minute timeout
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		log.Printf("Received message: %s\n", msg)

		metric, err := parseMetric(msg)
		if err != nil {
			log.Printf("%s - Error: %s", strings.Trim(msg, "\n"), err.Error())
			return
		}

		log.Printf("Received metric: %s\n", metric.Format("pretty"))
		for _, metricChan := range metrics {
			metricChan <- metric
		}
	}
}
