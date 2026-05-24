package main

import "github.com/prometheus/client_golang/prometheus"

var requestCount = prometheus.NewCounterVec(
	prometheus.CounterOpts{Name: "http_requests_total"},
	[]string{"method", "endpoint"},
)

func init() {
	prometheus.MustRegister(requestCount)
}
