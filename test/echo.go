package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func recordMetrics() {
	go func() {
		for {
			opsProcessed.Inc()
			time.Sleep(2 * time.Second)
		}
	}()
}

var (
	opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "myapp_processed_ops_total",
		Help: "The total number of processed events",
	})
)

func main() {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		recordMetrics()
		http.Handle("/metrics1", promhttp.Handler())
		http.ListenAndServe(":2112", nil)
		fmt.Printf("ListenAndServe :2112\n")
	}()

	go func() {
		defer wg.Done()
		recordMetrics()
		http.Handle("/metrics2", promhttp.Handler())
		http.ListenAndServe(":2113", nil)
		fmt.Printf("ListenAndServe :2113\n")
	}()

	wg.Wait()
}
