package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httptrace"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	timeMin = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "http_request_duration_seconds_min",
		Help: "Minimum time taken to perform the HTTP request in seconds",
	})
	timeMax = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "http_request_duration_seconds_max",
		Help: "Maximum time taken to perform the HTTP request in seconds",
	})
	timeAvg = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "http_request_duration_seconds_avg",
		Help: "Average time taken to perform the HTTP request in seconds",
	})
	totalRequests = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "http_request_total",
		Help: "Total number of HTTP requests performed",
	})
	responseStatus = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "http_request_status_code",
		Help: "HTTP response status code of the last request",
	})
)

func init() {
	prometheus.MustRegister(timeMin)
	prometheus.MustRegister(timeMax)
	prometheus.MustRegister(timeAvg)
	prometheus.MustRegister(totalRequests)
	prometheus.MustRegister(responseStatus)
}

func fetchGoogle(wg *sync.WaitGroup, times chan<- float64) {
	defer wg.Done()

	var localAddr string
	trace := &httptrace.ClientTrace{
		GotConn: func(connInfo httptrace.GotConnInfo) {
			localAddr = connInfo.Conn.LocalAddr().String()
		},
	}

	ctx := httptrace.WithClientTrace(context.Background(), trace)
	req, err := http.NewRequestWithContext(ctx, "GET", "http://sleep:8096", nil)
	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	elapsed := time.Since(start).Seconds()
	if elapsed > 0.5 {
		timestamp := time.Now().Format(time.RFC3339Nano)
		
		log.Printf("Slow request: %.2fms. Timestamp: %s. Source: %s\n", elapsed*1000, timestamp, localAddr)
	}

	if err != nil {
		log.Printf("Error al realizar la solicitud: %v", err)
		responseStatus.Set(0)
		return
	}
	defer resp.Body.Close()

	responseStatus.Set(float64(resp.StatusCode))

	times <- elapsed

}


func main() {
	nRequestsPerSecond := 200
	ticker := time.NewTicker(time.Second / time.Duration(nRequestsPerSecond))
	times := make(chan float64, nRequestsPerSecond*10) // Buffer para evitar bloqueos

	go func() {
		for {
			wg := &sync.WaitGroup{}
			for i := 0; i < nRequestsPerSecond; i++ {
				wg.Add(1)
				go fetchGoogle(wg, times)
				<-ticker.C
			}
			wg.Wait()

			close(times)
			sum := 0.0
			min := 1e9 
			max := 0.0
			count := 0

			for t := range times {
				sum += t
				if t < min {
					min = t
				}
				if t > max {
					max = t
				}
				count++
				totalRequests.Inc()
			}

			if count > 0 {
				timeAvg.Set(sum / float64(count))
				timeMin.Set(min)
				timeMax.Set(max)
			}
			times = make(chan float64, nRequestsPerSecond*10) 
		}
	}()

	http.Handle("/metrics", promhttp.Handler())

	log.Println("Server started at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
