// Must use separate test package to break import cycle.
package metrics_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/uber/jaeger-lib/metrics"
	"github.com/uber/jaeger-lib/metrics/metricstest"
)

func TestInitMetrics(t *testing.T) {
	testMetrics := struct {
		Gauge   metrics.Gauge   `metric:"gauge" tags:"1=one,2=two"`
		Counter metrics.Counter `metric:"counter"`
		Timer   metrics.Timer   `metric:"timer"`
	}{}

	f := metricstest.NewFactory(0)
	defer f.Stop()

	globalTags := map[string]string{"key": "value"}

	err := metrics.InitOrError(&testMetrics, f, globalTags)
	assert.NoError(t, err)

	testMetrics.Gauge.Update(10)
	testMetrics.Counter.Inc(5)
	testMetrics.Timer.Record(time.Duration(time.Second * 35))

	// wait for metrics
	for i := 0; i < 1000; i++ {
		c, _ := f.Snapshot()
		if _, ok := c["counter"]; ok {
			break
		}
		time.Sleep(1 * time.Millisecond)
	}

	c, g := f.Snapshot()

	assert.EqualValues(t, 5, c["counter|key=value"])
	assert.EqualValues(t, 10, g["gauge|1=one|2=two|key=value"])
	assert.EqualValues(t, 36863, g["timer|key=value.P50"])

	stopwatch := metrics.StartStopwatch(testMetrics.Timer)
	stopwatch.Stop()
	assert.True(t, 0 < stopwatch.ElapsedTime())
}

var (
	noMetricTag = struct {
		NoMetricTag metrics.Counter
	}{}

	badTags = struct {
		BadTags metrics.Counter `metric:"counter" tags:"1=one,noValue"`
	}{}

	invalidMetricType = struct {
		InvalidMetricType int64 `metric:"counter"`
	}{}
)

func TestInitMetricsFailures(t *testing.T) {
	assert.EqualError(t, metrics.InitOrError(&noMetricTag, nil, nil), "Field NoMetricTag is missing a tag 'metric'")

	assert.EqualError(t, metrics.InitOrError(&badTags, nil, nil),
		"Field [BadTags]: Tag [noValue] is not of the form key=value in 'tags' string [1=one,noValue]")

	assert.EqualError(t, metrics.InitOrError(&invalidMetricType, nil, nil),
		"Field InvalidMetricType is not a pointer to timer, gauge, or counter")
}

func TestInitPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("The code did not panic")
		}
	}()

	metrics.Init(&noMetricTag, metrics.NullFactory, nil)
}

func TestNullMetrics(t *testing.T) {
	// This test is just for cover
	metrics.NullFactory.Timer("name", nil).Record(0)
	metrics.NullFactory.Counter("name", nil).Inc(0)
	metrics.NullFactory.Gauge("name", nil).Update(0)
	metrics.NullFactory.Namespace("name", nil).Gauge("name2", nil).Update(0)
}
