package pushgateway

import (
	"fmt"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.k6.io/k6/metrics"
)

// CollectorResolver is an interface to resolve the various types of the [metrics.Metric]
// to the [prometheus.Collector].
// Respective [k6 metric type] are solved by following the [conversion rule] which
// the [xk6-output-prometheus-remote] extension applies.
//
// [k6 metric type]: https://k6.io/docs/using-k6/metrics/#metric-types
// [conversion rule]: https://k6.io/blog/k6-loves-prometheus/#mapping-k6-metrics-types
// [xk6-output-prometheus-remote]: https://github.com/grafana/xk6-output-prometheus-remote
type CollectorResolver func(metric *metrics.Metric, t time.Time) []prometheus.Collector

// CreateResolveer is a factory method to create the [ColloectorResolver] implementation
// corresponding to the given [k6 metric type].
//
// Example use case:
//
//    collectorResolver := collector_resolver.CreateResolver(sample.Metric.Type)
//    collectors := collectorResolver(sample.Metric, time.Now())
//
//
// [k6 metric type]: https://k6.io/docs/using-k6/metrics/#metric-types
func CreateResolver(t metrics.MetricType) CollectorResolver {
	var resolver CollectorResolver
	switch t {
	case metrics.Counter:
		resolver = resolveCounter
	case metrics.Gauge:
		resolver = resolveGauge
	case metrics.Rate:
		resolver = resolveRate
	case metrics.Trend:
		resolver = resolveTrend
	}
	return resolver
}

func resolveCounter(metric *metrics.Metric, t time.Time) []prometheus.Collector {
	counter := prometheus.NewCounterFunc(
		prometheus.CounterOpts{
			Name: metric.Name,
		},
		func() float64 {
			sink := metric.Sink.Format(time.Since(t))
			return sink["count"]
		},
	)
	return []prometheus.Collector{counter}
}

func resolveGauge(metric *metrics.Metric, t time.Time) []prometheus.Collector {
	gauge := prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: metric.Name,
		},
		func() float64 {
			sink := metric.Sink.Format(time.Since(t))
			return sink["value"]
		},
	)
	return []prometheus.Collector{gauge}
}

func resolveRate(metric *metrics.Metric, t time.Time) []prometheus.Collector {
	gauge := prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: metric.Name,
		},
		func() float64 {
			sink := metric.Sink.Format(time.Since(t))
			return sink["rate"]
		},
	)
	return []prometheus.Collector{gauge}
}

func resolveTrend(metric *metrics.Metric, t time.Time) []prometheus.Collector {
	sink := metric.Sink.Format(time.Since(t))

	collectors := make([]prometheus.Collector, 0)
	for k, v := range sink {
		// Remove "(" and ")" from the name of prometheus collector
		// Becuase these are not acceptable as collector name.
		suffix := strings.ReplaceAll(strings.ReplaceAll(k, "(", ""), ")", "")

		name := fmt.Sprintf("%s_%s", metric.Name, suffix)
		gauge := prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: name,
			},
		)
		gauge.Set(v)
		collectors = append(collectors, gauge)
	}
	return collectors
}
