package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"math"
	"math/rand"
	"net/http"
	"time"
)

var customLabels = []string{"app", "downstream"}

//var customLabels []string

var connections = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "latency",
	Help: "caculate connections by state",
}, customLabels)

func decimal(value float64) float64 {
	return math.Trunc(value*1e2+0.5) * 1e-2
}

func recordMetrics() {

	go func() {
		for {
			tikvValue := decimal(float64(rand.Intn(10)))
			connections.WithLabelValues("tidb", "tikv").Set(tikvValue)
			time.Sleep(2 * time.Second)
		}
	}()

	go func() {
		for {
			boltdbValue := decimal(float64(rand.Intn(10)))
			connections.WithLabelValues("tidb", "boltdb").Set(boltdbValue)
			time.Sleep(2 * time.Second)
		}
	}()
}

func main() {
	recordMetrics()
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2330", nil)
}
