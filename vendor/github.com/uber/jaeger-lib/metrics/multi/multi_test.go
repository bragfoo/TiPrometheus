package multi

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uber/jaeger-lib/metrics"
	"github.com/uber/jaeger-lib/metrics/metricstest"
)

var _ metrics.Factory = &Factory{} // API check

func TestMultiFactory(t *testing.T) {
	f1 := metricstest.NewFactory(time.Second)
	f2 := metricstest.NewFactory(time.Second)
	multi1 := New(f1, f2)
	multi2 := multi1.Namespace("ns2", nil)
	tags := map[string]string{"x": "y"}
	multi2.Counter("counter", tags).Inc(42)
	multi2.Gauge("gauge", tags).Update(42)
	multi2.Timer("timer", tags).Record(42 * time.Millisecond)

	for _, f := range []*metricstest.Factory{f1, f2} {
		f.AssertCounterMetrics(t,
			metricstest.ExpectedMetric{Name: "ns2.counter", Tags: tags, Value: 42})
		f.AssertGaugeMetrics(t,
			metricstest.ExpectedMetric{Name: "ns2.gauge", Tags: tags, Value: 42})
		_, g := f.Snapshot()
		assert.EqualValues(t, 43, g["ns2.timer|x=y.P99"])
	}
}
