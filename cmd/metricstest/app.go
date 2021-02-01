// Copyright 2021 The TiPrometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
